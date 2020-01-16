package claimsstore_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"testing"

	"github.com/dgrijalva/jwt-go"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/did/ethuri"
	"github.com/joincivil/id-hub/pkg/didjwt"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/testutils"
	didlib "github.com/ockam-network/did"
)

func TestAddJWT(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("failed to set up db connection")
	}
	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()

	didPersister := ethuri.NewPostgresPersister(db)
	ethURIRes := ethuri.NewService(didPersister)
	didService := did.NewService([]did.Resolver{ethURIRes})
	didJWTService := didjwt.NewService(didService)

	jwtClaimPersister := claimsstore.NewJWTClaimPGPersister(db, didJWTService)

	userDIDs := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4df7785c"
	userDID, err := didlib.Parse(userDIDs)

	if err != nil {
		t.Errorf("error parsing did: %v", err)
	}

	senderDIDs := "did:ethuri:e7ab0c43-d9fe-4a61-87a3-3fa99ce879e1"
	senderDID, err := didlib.Parse(senderDIDs)
	if err != nil {
		t.Errorf("error parsing did: %v", err)
	}

	secKey, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f")
	if err != nil {
		t.Errorf("error making ecdsa: %v", err)
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
		t.Errorf("error making the did doc: %v", err)
	}
	if err := ethURIRes.SaveDocument(didDoc); err != nil {
		t.Errorf("error saving the did doc: %v", err)
	}

	claims := &didjwt.VCClaimsJWT{
		Data: "",
		StandardClaims: jwt.StandardClaims{
			Issuer: "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4df7785c",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	tokenS, err := token.SignedString(secKey)

	if err != nil {
		t.Errorf("error creating token string: %v", err)
	}

	hash, err := jwtClaimPersister.AddJWT(tokenS, senderDID)
	if err != nil {
		t.Errorf("error adding the jwt: %v", err)
	}

	retrievedToken, err := jwtClaimPersister.GetJWTByMultihash(hash)
	if err != nil {
		t.Errorf("error retrieving jwt by hash: %v", err)
	}

	claims, ok := retrievedToken.Claims.(*didjwt.VCClaimsJWT)

	if !ok {
		t.Errorf("invlaid claims type")
	}

	if claims.Issuer != "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4df7785c" {
		t.Errorf("wrong issuer returned from token")
	}
}
