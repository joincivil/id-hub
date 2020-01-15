package didjwt_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/didjwt"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/testutils"
	didlib "github.com/ockam-network/did"
)

func setupConnection() (*gorm.DB, error) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		return nil, err
	}
	db.DropTable(&did.PostgresDocument{}, &claimsstore.RootCommit{}, &claimsstore.Node{})
	err = db.AutoMigrate(&did.PostgresDocument{}, &claimsstore.SignedClaimPostgres{}, &claimsstore.Node{}, &claimsstore.RootCommit{}).Error
	if err != nil {
		return nil, err
	}

	return db, nil
}

func addPubKey(didService *did.Service, pubKey *ecdsa.PublicKey, doc *did.Document) error {
	pubBytes := crypto.FromECDSAPub(pubKey)
	pub := hex.EncodeToString(pubBytes)
	docPubKey := &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeSecp256r1Verification,
		PublicKeyHex: &pub,
	}

	docPubKey.ID = did.CopyDID(&doc.ID)
	docPubKey.Controller = did.CopyDID(&doc.ID)

	err := doc.AddPublicKey(docPubKey, false, true)
	if err != nil {
		return err
	}

	if err := didService.SaveDocument(doc); err != nil {
		return err
	}
	return nil
}

func TestParseJWT(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("failed to set up db connection")
	}
	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()

	didPersister := did.NewPostgresPersister(db)
	didService := did.NewService(didPersister)

	didJWTService := didjwt.NewService(didService)

	userDIDs := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"
	userDID, err := didlib.Parse(userDIDs)
	if err != nil {
		t.Errorf("error parsing did: %v", err)
	}
	privateKey, err := crypto.HexToECDSA("79156abe7fe2fd433dc9df969286b96666489bac508612d0e16593e944c4f69f") // key generated just for tests

	if err != nil {
		t.Errorf("error making ecdsa: %v", err)
	}
	pubKey := privateKey.Public().(*ecdsa.PublicKey)

	pubBytes := crypto.FromECDSAPub(pubKey)
	pub := hex.EncodeToString(pubBytes)
	docPubKey := &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: &pub,
	}

	docPubKey.ID = did.CopyDID(userDID)
	docPubKey.Controller = did.CopyDID(userDID)
	didDoc, err := did.InitializeNewDocument(userDID, docPubKey, true, true)
	if err != nil {
		t.Errorf("error making the did doc: %v", err)
	}
	if err := didService.SaveDocument(didDoc); err != nil {
		t.Errorf("error saving the did doc: %v", err)
	}

	privateKey2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("error making ecdsa: %v", err)
	}
	pubKey2 := privateKey2.Public().(*ecdsa.PublicKey)

	err = addPubKey(didService, pubKey2, didDoc)
	if err != nil {
		t.Errorf("error adding pubkey: %v", err)
	}

	privateKey3, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("error making ecdsa: %v", err)
	}
	pubKey3 := privateKey3.Public().(*ecdsa.PublicKey)

	err = addPubKey(didService, pubKey3, didDoc)
	if err != nil {
		t.Errorf("error adding pubkey: %v", err)
	}

	claims := &didjwt.VCClaimsJWT{
		Data: "",
		StandardClaims: jwt.StandardClaims{
			Issuer: "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	tokenS, err := token.SignedString(privateKey2)

	if err != nil {
		t.Errorf("error creating token string: %v", err)
	}

	parsedToken, err := didJWTService.ParseJWT(tokenS)

	if err != nil {
		t.Errorf("could not verify the token: %v", err)
	}

	if !parsedToken.Valid {
		t.Errorf("token should be valid")
	}

	claims, ok := parsedToken.Claims.(*didjwt.VCClaimsJWT)

	if !ok {
		t.Errorf("invlaid claims type")
	}

	if claims.Issuer != "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c" {
		t.Errorf("wrong issuer returned from token")
	}

	privateKey4, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("error making ecdsa: %v", err)
	}

	tokenS, err = token.SignedString(privateKey4)
	if err != nil {
		t.Errorf("error creating token string: %v", err)
	}

	_, err = didJWTService.ParseJWT(tokenS)
	if err == nil {
		t.Errorf("should not have been able to verify a jwt made with pk not assigned to did")
	}

}
