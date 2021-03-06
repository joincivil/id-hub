package claims_test

import (
	"encoding/hex"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/multiformats/go-multihash"
	didlib "github.com/ockam-network/did"

	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/didjwt"
	"github.com/joincivil/id-hub/pkg/testinits"
	"github.com/joincivil/id-hub/pkg/testutils"
)

func TestJWTService(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	didService, ethURI := testinits.InitDIDService(db)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, rootService, err := testinits.MakeService(db, didService, signedClaimStore)
	if err != nil {
		t.Errorf("error setting up service: %v", err)
	}

	didJWTService := didjwt.NewService(didService)

	jwtClaimPersister := claimsstore.NewJWTClaimPGPersister(db, didJWTService)

	jwtService := claims.NewJWTService(didJWTService, jwtClaimPersister, claimService, &testutils.FakePubSubService{})

	senderDIDs := "did:ethuri:e7ab0c43-d9fe-4a61-87a3-3fa99ce879e1"
	senderDID, _ := didlib.Parse(senderDIDs)

	userDID, secKey, err := testinits.AddDID(ethURI, claimService)

	if err != nil {
		t.Errorf("failed to add userdid: %v", err)
	}

	err = rootService.CommitRoot()
	if err != nil {
		t.Errorf("error committing root: %v", err)
	}

	claims := &didjwt.VCClaimsJWT{
		Data: "",
		StandardClaims: jwt.StandardClaims{
			Issuer: userDID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	tokenS, err := token.SignedString(secKey)
	if err != nil {
		t.Errorf("unable to create jwt string: %v", err)
	}

	_, err = jwtService.AddJWTClaim(tokenS, senderDID)

	if err != nil {
		t.Errorf("failed to add jwt: %v", err)
	}

	usersTokens, err := jwtService.GetJWTSforDID(userDID)

	if err != nil {
		t.Errorf("failed to fetch tokens: %v", err)
	}

	if len(usersTokens) != 1 {
		t.Errorf("wrong number of tokens returned got: %v, expected: %v", len(usersTokens), 1)
	}

	proofBeforeCommit, err := jwtService.GenerateProof(tokenS)
	if err != nil {
		t.Errorf("error generating proof: %v", err)
	}

	if proofBeforeCommit.BlockNumber != -1 {
		t.Errorf("block number should be -1 before committing the root")
	}

	err = rootService.CommitRoot()
	if err != nil {
		t.Errorf("error committing root: %v", err)
	}

	proof, err := jwtService.GenerateProof(tokenS)
	if err != nil {
		t.Errorf("error generating proof: %v", err)
	}

	if proof.BlockNumber != 3 {
		t.Errorf("didn't get the right blocknumber")
	}

	existsProofBytes, err := hex.DecodeString(proof.ExistsInDIDMTProof)
	if err != nil {
		t.Errorf("couldn't decode exists proof: %v", err)
	}
	notRevokedBytes, err := hex.DecodeString(proof.NotRevokedInDIDMTProof)
	if err != nil {
		t.Errorf("couldn't decode nonrevoked proof: %v", err)
	}
	didRootBytes, err := hex.DecodeString(proof.DIDRootExistsProof)
	if err != nil {
		t.Errorf("couldn't decode didroot proof: %v", err)
	}

	existProof, err := merkletree.NewProofFromBytes(existsProofBytes)
	if err != nil {
		t.Errorf("couldn't build exists proof from bytes: %v", err)
	}
	notRevoked, err := merkletree.NewProofFromBytes(notRevokedBytes)
	if err != nil {
		t.Errorf("couldn't build not revoked proof from bytes: %v", err)
	}
	didRoot, err := merkletree.NewProofFromBytes(didRootBytes)
	if err != nil {
		t.Errorf("couldn't build did root proof from bytes: %v", err)
	}

	if err != nil {
		t.Errorf("couldn't get root of did tree: %v", err)
	}

	hash := crypto.Keccak256([]byte(tokenS))
	mhash, _ := multihash.EncodeName(hash, "keccak-256")
	hash34 := [34]byte{}
	copy(hash34[:], mhash)

	rdClaim, _ := claimtypes.NewClaimRegisteredDocument(hash34, userDID, claimtypes.JWTDocType)

	entry := rdClaim.Entry()
	if !merkletree.VerifyProof(&proof.DIDRoot, existProof, entry.HIndex(), entry.HValue()) {
		t.Errorf("couldn't verify exists in did tree proof")
	}
	// revoked registered doc claim is always version 1
	rdClaim.Version = 1
	entryV1 := rdClaim.Entry()
	if !merkletree.VerifyProof(&proof.DIDRoot, notRevoked, entryV1.HIndex(), entryV1.HValue()) {
		t.Errorf("couldn't verify not revoked in proof")
	}

	rootClaim, _ := claimtypes.NewClaimSetRootKeyDID(userDID, &proof.DIDRoot)
	rootClaim.Version = proof.DIDRootExistsVersion
	rootClaimEntry := rootClaim.Entry()

	if !merkletree.VerifyProof(&proof.Root, didRoot, rootClaimEntry.HIndex(), rootClaimEntry.HValue()) {
		t.Errorf("couldn't verify root tree proof")
	}

	err = jwtService.RevokeJWTClaim(tokenS)
	if err != nil {
		t.Errorf("couldn't revoke claim")
	}

	_, err = jwtService.GenerateProof(tokenS)
	if err == nil {
		t.Errorf("it should error if the claim is revoked")
	}

}
