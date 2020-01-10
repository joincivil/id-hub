package did_test

import (
	"fmt"
	"testing"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/testutils"
	"github.com/joincivil/id-hub/pkg/utils"
	didlib "github.com/ockam-network/did"
)

const (
	testKeyVal = "046539bd140ab14032735641692cbc3e7b52ef9e367887f4f2fd53942c870a5279c8639a511d9965c56c13fc7b00e636ecf0ea77237dd3e363a31ce95a06e58080"
)

func TestKeyFromType(t *testing.T) {
	pkey := "04f3df3cea421eac2a7f5dbd8e8d505470d42150334f512bd6383c7dc91bf8fa4d5458d498b4dcd05574c902fb4c233005b3f5f3ff3904b41be186ddbda600580b"
	pk := &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: &pkey,
	}
	key, err := did.KeyFromType(pk)
	if err != nil {
		t.Errorf("Should not have returned an error: err: %v", err)
	}
	if *key != pkey {
		t.Errorf("Should have returned the hex field")
	}

	pk = &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyJwk: &pkey,
	}
	key, err = did.KeyFromType(pk)
	if err == nil {
		t.Errorf("Should have returned an error: err: %v", err)
	}
	if key != nil {
		t.Errorf("Should have returned a nil key")
	}

	pk = &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeRsaVerification,
		PublicKeyHex: &pkey,
	}
	key, err = did.KeyFromType(pk)
	if err == nil {
		t.Errorf("Should have returned an error: err: %v", err)
	}
	if key != nil {
		t.Errorf("Should have returned a nil key")
	}
}

func TestValidDocPublicKey(t *testing.T) {
	d, _ := didlib.Parse("did:ethuri:123456#keys-1")
	controller1, _ := didlib.Parse("did:ethuri:123456")

	valid := did.ValidDocPublicKey(&did.DocPublicKey{
		ID:           d,
		Controller:   controller1,
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: utils.StrToPtr(testKeyVal),
	})
	if !valid {
		t.Errorf("Should have been a valid key")
	}

	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		ID:           d,
		Controller:   controller1,
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: utils.StrToPtr("thisisinvalid"),
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}

	// malformed public hex key value
	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		ID:           d,
		Controller:   controller1,
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: utils.StrToPtr("046539bd140ab14032735641692cbc3e7b52ef9e367887f4f2fd53942c870a5279c8639a511d9965c56c13fc7b00e636ecf0ea77237dd3e363a31ce95a06e58081"),
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}

	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		ID:           d,
		Controller:   controller1,
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: utils.StrToPtr(""),
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}

	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		ID:           d,
		Controller:   controller1,
		Type:         linkeddata.SuiteTypeEd25519Signature,
		PublicKeyHex: utils.StrToPtr(testKeyVal),
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}

	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		Type:         linkeddata.SuiteTypeEd25519Signature,
		PublicKeyHex: utils.StrToPtr(testKeyVal),
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}

	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		Controller:   controller1,
		Type:         linkeddata.SuiteTypeEd25519Signature,
		PublicKeyHex: utils.StrToPtr(testKeyVal),
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}
}

func TestPublicKeyInSlice(t *testing.T) {
	doc := testutils.BuildTestDocument()
	testPk := doc.PublicKeys[0]
	if !did.PublicKeyInSlice(testPk, doc.PublicKeys) {
		t.Errorf("Should have found public key in slice")
	}

	pk1 := did.DocPublicKey{}
	pk1ID := fmt.Sprintf("%v#keys-3", doc.ID.String())
	d1, _ := didlib.Parse(pk1ID)
	pk1.ID = d1
	pk1.Type = linkeddata.SuiteTypeSecp256k1Verification
	pk1.Controller = did.CopyDID(&doc.ID)
	// different key
	hexKey := "04ad8439b0cc03a2f45504b4c7ec68c5c6372da7322071e5828d78205be417d1c9876e8797b2d1e2211fbfaf434ac0c421e1a703e55dc9cd2e024e6b462cc9e0ee"
	pk1.PublicKeyHex = utils.StrToPtr(hexKey)

	if did.PublicKeyInSlice(pk1, doc.PublicKeys) {
		t.Errorf("Should not have found public key in slice")
	}
}

func TestAuthInSlice(t *testing.T) {
	doc := testutils.BuildTestDocument()

	testAuth := doc.Authentications[0]
	if !did.AuthInSlice(testAuth, doc.Authentications) {
		t.Errorf("Should have found auth in slice")
	}

	aw2 := did.DocAuthenicationWrapper{}
	aw2ID := fmt.Sprintf("%v#keys-3", doc.ID.String())
	d4, _ := didlib.Parse(aw2ID)
	aw2.ID = d4
	aw2.IDOnly = false
	aw2.Type = linkeddata.SuiteTypeSecp256k1Verification
	aw2.Controller = did.CopyDID(&doc.ID)
	hexKey2 := "04ad8439b0cc03a2f45504b4c7ec68c5c6372da7322071e5828d78205be417d1c9876e8797b2d1e2211fbfaf434ac0c421e1a703e55dc9cd2e024e6b462cc9e0ee"
	aw2.PublicKeyHex = utils.StrToPtr(hexKey2)

	if did.AuthInSlice(aw2, doc.Authentications) {
		t.Errorf("Should not have found auth in slice")
	}
}

func TestServiceInSlice(t *testing.T) {
	doc := testutils.BuildTestDocument()
	testSrv := doc.Services[0]
	if !did.ServiceInSlice(testSrv, doc.Services) {
		t.Errorf("Should have found service in slice")
	}

	ep1 := did.DocService{}
	ep1ID := fmt.Sprintf("%v#vcr", doc.ID.String())
	d2, _ := didlib.Parse(ep1ID)
	ep1.ID = *d2
	ep1.Type = "IdentityHub"
	ep1.PublicKey = "did:example:123456789abcdefghi#key-1"
	ep1.ServiceEndpointLD = map[string]interface{}{
		"@context":  "https://schema.identity.foundation/hub",
		"type":      "UserEndPoint",
		"instances": []string{"did:example:456", "did:example:789"},
	}
	ep1.ServiceEndpoint = ep1.ServiceEndpointLD

	if did.ServiceInSlice(ep1, doc.Services) {
		t.Errorf("Should not have found service in slice")
	}
}
