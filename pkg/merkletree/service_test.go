package merkletree_test

import (
	"encoding/hex"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/did/ethuri"
	"github.com/joincivil/id-hub/pkg/didjwt"
	mt "github.com/joincivil/id-hub/pkg/merkletree"
	"github.com/joincivil/id-hub/pkg/testinits"
	"github.com/joincivil/id-hub/pkg/testutils"
	"github.com/joincivil/id-hub/pkg/utils"
)

func TestMerkletreeService(t *testing.T) {
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

	merkleTreeService := mt.NewService(didJWTService, claimService)

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

	_, err = merkleTreeService.AddEntry(tokenS)

	if err != nil {
		t.Errorf("failed to add jwt: %v", err)
	}

	proofBeforeCommit, err := merkleTreeService.GenerateProof(tokenS)
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

	proof, err := merkleTreeService.GenerateProof(tokenS)
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

	mhash, _ := utils.CreateMultihash([]byte(tokenS))
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

	err = merkleTreeService.RevokeEntry(tokenS)
	if err != nil {
		t.Errorf("couldn't revoke claim")
	}

	_, err = merkleTreeService.GenerateProof(tokenS)
	if err == nil {
		t.Errorf("it should error if the claim is revoked")
	}
}

func setupConnection() (*gorm.DB, error) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		return nil, err
	}
	db.DropTable(&ethuri.PostgresDocument{}, &claimsstore.RootCommit{}, &claimsstore.Node{})
	err = db.AutoMigrate(&ethuri.PostgresDocument{}, &claimsstore.SignedClaimPostgres{}, &claimsstore.Node{}, &claimsstore.RootCommit{}, &claimsstore.JWTClaimPostgres{}).Error
	if err != nil {
		return nil, err
	}
	return db, nil
}
