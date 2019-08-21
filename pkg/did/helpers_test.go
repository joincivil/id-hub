package did_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/joincivil/id-hub/pkg/did"
)

const (
	testKeyVal = "046539bd140ab14032735641692cbc3e7b52ef9e367887f4f2fd53942c870a5279c8639a511d9965c56c13fc7b00e636ecf0ea77237dd3e363a31ce95a06e58080"
)

func TestGenerateNewDocument(t *testing.T) {
	pk := &did.DocPublicKey{
		Type:         did.LDSuiteTypeSecp256k1Verification,
		PublicKeyHex: testKeyVal,
	}

	doc, err := did.GenerateNewDocument(pk)
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
		PublicKeyHex: testKeyVal,
	}
	newDoc, err := did.InitializeNewDocument(newDID, firstPK)
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

func TestValidDocPublicKey(t *testing.T) {
	valid := did.ValidDocPublicKey(&did.DocPublicKey{
		Type:         did.LDSuiteTypeSecp256k1Verification,
		PublicKeyHex: testKeyVal,
	})
	if !valid {
		t.Errorf("Should have been a valid key")
	}

	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		Type:         did.LDSuiteTypeSecp256k1Verification,
		PublicKeyHex: "thisisinvalid",
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}

	// malformed public hex key value
	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		Type:         did.LDSuiteTypeSecp256k1Verification,
		PublicKeyHex: "046539bd140ab14032735641692cbc3e7b52ef9e367887f4f2fd53942c870a5279c8639a511d9965c56c13fc7b00e636ecf0ea77237dd3e363a31ce95a06e58081",
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}

	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		Type:         did.LDSuiteTypeSecp256k1Verification,
		PublicKeyHex: "",
	})
	if valid {
		t.Errorf("Should have been a invalid key")
	}

	valid = did.ValidDocPublicKey(&did.DocPublicKey{
		Type:         did.LDSuiteTypeEd25519Signature,
		PublicKeyHex: testKeyVal,
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

	if docPK.PublicKeyHex != testKeyVal {
		t.Errorf("should have been set to hex field")
	}

}
