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

func makeDID(id *didlib.DID, key *ecdsa.PublicKey, ethURIRes *ethuri.Service) error {
	pubBytes := crypto.FromECDSAPub(key)
	pub := hex.EncodeToString(pubBytes)
	docPubKey := &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: &pub,
	}

	docPubKey.ID = did.CopyDID(id)
	docPubKey.Controller = did.CopyDID(id)
	didDoc, err := ethuri.InitializeNewDocument(id, docPubKey, false, true)
	if err != nil {
		return err
	}
	if err := ethURIRes.SaveDocument(didDoc); err != nil {
		return err
	}
	return nil
}

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

	err = makeDID(userDID, pubKey, ethURIRes)
	if err != nil {
		t.Errorf("error creating did doc: %v", err)
	}

	secKey2, err := crypto.HexToECDSA("6639112124e9903c6ca9397078a1508bb333a71908b0c467c14aa5882dcb59a8")
	if err != nil {
		t.Errorf("error making private key: %v", err)
	}

	pubKey = secKey2.Public().(*ecdsa.PublicKey)

	err = makeDID(senderDID, pubKey, ethURIRes)
	if err != nil {
		t.Errorf("error creating did doc: %v", err)
	}

	claims := &didjwt.VCClaimsJWT{
		Data: "",
		StandardClaims: jwt.StandardClaims{
			Issuer:  "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4df7785c",
			Subject: "did:ethuri:abcdefghijklmnopqrstuvwxyz",
		},
	}

	claim2 := &didjwt.VCClaimsJWT{
		Data: "",
		StandardClaims: jwt.StandardClaims{
			Issuer:  "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4df7785c",
			Subject: "did:ethuri:zyxwt",
		},
	}

	claim3 := &didjwt.VCClaimsJWT{
		Data: "",
		StandardClaims: jwt.StandardClaims{
			Issuer:  "did:ethuri:e7ab0c43-d9fe-4a61-87a3-3fa99ce879e1",
			Subject: "did:ethuri:abcdefghijklmnopqrstuvwxyz",
		},
	}

	claim4 := &didjwt.VCClaimsJWT{
		Data: "",
		StandardClaims: jwt.StandardClaims{
			Issuer:  "did:ethuri:e7ab0c43-d9fe-4a61-87a3-3fa99ce879e1",
			Subject: "did:ethuri:zyxwt",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	tokenS, err := token.SignedString(secKey)

	if err != nil {
		t.Errorf("error creating token string: %v", err)
	}

	_, hash, err := jwtClaimPersister.AddJWT(tokenS, senderDID)
	if err != nil {
		t.Errorf("error adding the jwt: %v", err)
	}

	token = jwt.NewWithClaims(jwt.SigningMethodES256, claim2)

	tokenS, err = token.SignedString(secKey)

	if err != nil {
		t.Errorf("error creating token string: %v", err)
	}

	_, _, err = jwtClaimPersister.AddJWT(tokenS, senderDID)
	if err != nil {
		t.Errorf("error adding the jwt: %v", err)
	}

	token = jwt.NewWithClaims(jwt.SigningMethodES256, claim3)

	tokenS, err = token.SignedString(secKey2)

	if err != nil {
		t.Errorf("error creating token string: %v", err)
	}

	_, _, err = jwtClaimPersister.AddJWT(tokenS, senderDID)
	if err != nil {
		t.Errorf("error adding the jwt: %v", err)
	}

	token = jwt.NewWithClaims(jwt.SigningMethodES256, claim4)

	tokenS, err = token.SignedString(secKey2)

	if err != nil {
		t.Errorf("error creating token string: %v", err)
	}

	_, _, err = jwtClaimPersister.AddJWT(tokenS, senderDID)
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

	retrievedTokens, err := jwtClaimPersister.GetJWTBySubjectsOrIssuers([]string{"did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4df7785c"}, []string{"did:ethuri:abcdefghijklmnopqrstuvwxyz"})

	if err != nil {
		t.Errorf("error retrieving tokens: %v", err)
	}

	if len(retrievedTokens) != 3 {
		t.Errorf("wrong number of tokens returned")
	}
}
