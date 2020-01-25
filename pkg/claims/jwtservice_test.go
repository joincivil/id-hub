package claims_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/crypto"
	didlib "github.com/ockam-network/did"

	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/did/ethuri"
	"github.com/joincivil/id-hub/pkg/didjwt"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/testutils"
)

func addDID(ethURI *ethuri.Service) (*didlib.DID, *ecdsa.PrivateKey, error) {
	userDIDs := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4df7785c"
	userDID, err := didlib.Parse(userDIDs)

	if err != nil {
		return nil, nil, err
	}

	secKey, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		return nil, nil, err
	}
	pubKey := secKey.Public().(*ecdsa.PublicKey)

	pubBytes := crypto.FromECDSAPub(pubKey)
	pub := hex.EncodeToString(pubBytes)
	docPubKey := &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: &pub,
	}

	docPubKey.ID = did.CopyDID(userDID)
	docPubKey.Controller = did.CopyDID(userDID)
	didDoc, err := ethuri.InitializeNewDocument(userDID, docPubKey, false, true)
	if err != nil {
		return nil, nil, err
	}
	if err := ethURI.SaveDocument(didDoc); err != nil {
		return nil, nil, err
	}

	return userDID, secKey, nil

}

func TestJWTService(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	didService, ethURI := initDIDService(db)
	signedClaimStore := claimsstore.NewSignedClaimPGPersister(db)
	claimService, _, err := makeService(db, didService, signedClaimStore)
	if err != nil {
		t.Errorf("error setting up service: %v", err)
	}
	didJWTService := didjwt.NewService(didService)

	jwtClaimPersister := claimsstore.NewJWTClaimPGPersister(db, didJWTService)

	jwtService := claims.NewJWTService(didJWTService, jwtClaimPersister, claimService)

	senderDIDs := "did:ethuri:e7ab0c43-d9fe-4a61-87a3-3fa99ce879e1"
	senderDID, _ := didlib.Parse(senderDIDs)

	userDID, secKey, err := addDID(ethURI)

	if err != nil {
		t.Errorf("failed to add userdid: %v", err)
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

	err = jwtService.AddJWTClaim(tokenS, senderDID)

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

}
