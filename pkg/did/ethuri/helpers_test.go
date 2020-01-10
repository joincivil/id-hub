package ethuri_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/did/ethuri"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/utils"
)

const (
	testKeyVal = "046539bd140ab14032735641692cbc3e7b52ef9e367887f4f2fd53942c870a5279c8639a511d9965c56c13fc7b00e636ecf0ea77237dd3e363a31ce95a06e58080"
)

func TestGenerateNewDocument(t *testing.T) {
	pk := &did.DocPublicKey{
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		PublicKeyHex: utils.StrToPtr(testKeyVal),
	}

	doc, err := ethuri.GenerateNewDocument(pk, true, true)
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
	newDID, err := ethuri.GenerateEthURIDID()
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
	newDID, err := ethuri.GenerateEthURIDID()
	if err != nil {
		t.Errorf("Should have not gotten error on DID generation")
	}

	firstPK := &did.DocPublicKey{
		ID:           did.CopyDID(newDID),
		Type:         linkeddata.SuiteTypeSecp256k1Verification,
		Controller:   did.CopyDID(newDID),
		PublicKeyHex: utils.StrToPtr(string(testKeyVal)),
	}
	newDoc, err := ethuri.InitializeNewDocument(newDID, firstPK, true, true)
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
