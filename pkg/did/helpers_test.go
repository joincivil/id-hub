package did_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/utils"
	didlib "github.com/ockam-network/did"
)

const (
	testKeyVal = "046539bd140ab14032735641692cbc3e7b52ef9e367887f4f2fd53942c870a5279c8639a511d9965c56c13fc7b00e636ecf0ea77237dd3e363a31ce95a06e58080"
)

func TestGenerateNewDocument(t *testing.T) {
	pk := &did.DocPublicKey{
		Type:         did.LDSuiteTypeSecp256k1Verification,
		PublicKeyHex: utils.StrToPtr(testKeyVal),
	}

	doc, err := did.GenerateNewDocument(pk, true, true)
	if err != nil {
		t.Fatalf("Should not have returned error for generating a new doc: err: %v", err)
	}

	if doc.ID.String() == "" {
		t.Errorf("Should have generated a new DID")
	}

	if len(doc.PublicKeys) != 1 {
		t.Errorf("Should have setup one public key")
	}
	if len(doc.Authentications) != 1 {
		t.Errorf("Should have setup one authentication")
	}
}

func TestGenerateEthURIDID(t *testing.T) {
	newDID, err := did.GenerateEthURIDID()
	if err != nil {
		t.Errorf("Should have not gotten error on DID generation")
	}
	if newDID == nil {
		t.Errorf("Should have not gotten nil DID")
	}
	if !strings.HasPrefix(newDID.String(), "did:ethuri:") {
		t.Errorf("Should have gotten did:ethuri: prefix, '%v'", newDID.String())
	}
}

func TestInitializeNewDocument(t *testing.T) {
	newDID, err := did.GenerateEthURIDID()
	if err != nil {
		t.Errorf("Should have not gotten error on DID generation")
	}

	firstPK := &did.DocPublicKey{
		ID:           did.CopyDID(newDID),
		Type:         did.LDSuiteTypeSecp256k1Verification,
		Controller:   did.CopyDID(newDID),
		PublicKeyHex: utils.StrToPtr(string(testKeyVal)),
	}
	newDoc, err := did.InitializeNewDocument(newDID, firstPK, true, true)
	if err != nil {
		t.Errorf("Should not have gotten error generating new doc")
	}

	if newDoc.ID.String() != newDID.String() {
		t.Errorf("Should have gotten same DID")
	}

	if len(newDoc.PublicKeys) != 1 {
		t.Errorf("Should have gotten 1 key")
	}

	pk := newDoc.PublicKeys[0]
	if pk.ID.String() != fmt.Sprintf("%v#keys-1", newDID.String()) {
		t.Errorf("Should have gotten key-1 in public key id")
	}

	if len(newDoc.Authentications) != 1 {
		t.Errorf("Should have gotten 1 authentication")
	}
	auth := newDoc.Authentications[0]
	if !auth.IDOnly {
		t.Errorf("Should have gotten ID only")
	}
	if auth.ID.String() != fmt.Sprintf("%v#keys-1", newDID.String()) {
		t.Errorf("Should have gotten key-1 in public key id")
	}

	bys, _ := json.MarshalIndent(newDoc, "", "    ")
	t.Logf("%v", string(bys))
}

func TestCopyDID(t *testing.T) {
	d, _ := did.GenerateEthURIDID()
	cpy := did.CopyDID(d)
	if cpy.String() != d.String() {
		t.Errorf("Should have matching DID strings")
	}
	if cpy == d {
		t.Errorf("Should not be the same exact struct value in mem")
	}
}

func TestValidDid(t *testing.T) {
	if did.ValidDid("notavaliddid") {
		t.Errorf("Should not have returned true as valid did")
	}
	if did.ValidDid("") {
		t.Errorf("Should not have returned true as valid did")
	}
	if did.ValidDid("uri:123345") {
		t.Errorf("Should not have returned true as valid did")
	}
	if !did.ValidDid("did:uri:123345") {
		t.Errorf("Should have returned true as valid did")
	}
}

func TestValidDocPublicKey(t *testing.T) {
	valid := did.ValidDocPublicKey(&did.DocPublicKey{
		Type:         did.LDSuiteTypeSecp256k1Verification,
		PublicKeyHex: utils.StrToPtr(testKeyVal),
	})
	if !valid {
		t.Errorf("Should have been a valid key")
	}

	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		Type:         did.LDSuiteTypeSecp256k1Verification,
		PublicKeyHex: utils.StrToPtr("thisisinvalid"),
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}

	// malformed public hex key value
	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		Type:         did.LDSuiteTypeSecp256k1Verification,
		PublicKeyHex: utils.StrToPtr("046539bd140ab14032735641692cbc3e7b52ef9e367887f4f2fd53942c870a5279c8639a511d9965c56c13fc7b00e636ecf0ea77237dd3e363a31ce95a06e58081"),
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}

	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		Type:         did.LDSuiteTypeSecp256k1Verification,
		PublicKeyHex: utils.StrToPtr(""),
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}

	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		Type:         did.LDSuiteTypeEd25519Signature,
		PublicKeyHex: utils.StrToPtr(testKeyVal),
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}
}

func TestValidateBuildDocPublicKey(t *testing.T) {
	docPK := did.ValidateBuildDocPublicKey(
		did.LDSuiteTypeSecp256k1Verification,
		testKeyVal,
	)

	if docPK == nil {
		t.Errorf("should not have received nil doc public key")
	}

	if docPK.Type != did.LDSuiteTypeSecp256k1Verification {
		t.Errorf("should have been set to secp251k verification")
	}

	if *docPK.PublicKeyHex != testKeyVal {
		t.Errorf("should have been set to hex field")
	}

}

func TestPublicKeyInSlice(t *testing.T) {
	doc := did.BuildTestDocument()
	testPk := doc.PublicKeys[0]
	if !did.PublicKeyInSlice(testPk, doc.PublicKeys) {
		t.Errorf("Should have found public key in slice")
	}

	pk1 := did.DocPublicKey{}
	pk1ID := fmt.Sprintf("%v#keys-3", doc.ID.String())
	d1, _ := didlib.Parse(pk1ID)
	pk1.ID = d1
	pk1.Type = did.LDSuiteTypeSecp256k1Verification
	pk1.Controller = did.CopyDID(&doc.ID)
	// different key
	hexKey := "04ad8439b0cc03a2f45504b4c7ec68c5c6372da7322071e5828d78205be417d1c9876e8797b2d1e2211fbfaf434ac0c421e1a703e55dc9cd2e024e6b462cc9e0ee"
	pk1.PublicKeyHex = utils.StrToPtr(hexKey)

	if did.PublicKeyInSlice(pk1, doc.PublicKeys) {
		t.Errorf("Should not have found public key in slice")
	}
}

func TestAuthInSlice(t *testing.T) {
	doc := did.BuildTestDocument()

	testAuth := doc.Authentications[0]
	if !did.AuthInSlice(testAuth, doc.Authentications) {
		t.Errorf("Should have found auth in slice")
	}

	aw2 := did.DocAuthenicationWrapper{}
	aw2ID := fmt.Sprintf("%v#keys-3", doc.ID.String())
	d4, _ := didlib.Parse(aw2ID)
	aw2.ID = d4
	aw2.IDOnly = false
	aw2.Type = did.LDSuiteTypeSecp256k1Verification
	aw2.Controller = did.CopyDID(&doc.ID)
	hexKey2 := "04ad8439b0cc03a2f45504b4c7ec68c5c6372da7322071e5828d78205be417d1c9876e8797b2d1e2211fbfaf434ac0c421e1a703e55dc9cd2e024e6b462cc9e0ee"
	aw2.PublicKeyHex = utils.StrToPtr(hexKey2)

	if did.AuthInSlice(aw2, doc.Authentications) {
		t.Errorf("Should not have found auth in slice")
	}
}

func TestServiceInSlice(t *testing.T) {
	doc := did.BuildTestDocument()
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