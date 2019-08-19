package did_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/testutils"
)

func initPersister(t *testing.T) (did.Persister, *gorm.DB) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		t.Fatalf("Should have returned a new gorm db conn")
		return nil, nil
	}

	err = db.AutoMigrate(&did.PostgresDocument{}).Error
	if err != nil {
		t.Fatalf("Should have auto-migrated")
		return nil, nil
	}

	return did.NewPostgresPersister(db), db
}

func initService(t *testing.T) (*did.Service, *gorm.DB) {
	persister, db := initPersister(t)
	return did.NewService(persister), db
}

func TestServiceSaveGetDocument(t *testing.T) {
	service, db := initService(t)
	defer db.DropTable(&did.PostgresDocument{})

	doc := did.BuildTestDocument()

	// Save document
	err := service.SaveDocument(doc)
	if err != nil {
		t.Errorf("Should have not gotten error saving document: err: %v", err)
	}

	// Get document
	doc, err = service.GetDocument(doc.ID.String())
	if err != nil {
		t.Errorf("Should have not gotten error retrieving document: err: %v", err)
	}
	if doc == nil {
		t.Errorf("Should have not gotten nil document: err: %v", err)
	}

	// Get document via a did
	doc, err = service.GetDocumentFromDID(&doc.ID)
	if err != nil {
		t.Errorf("Should have not gotten error retrieving document: err: %v", err)
	}
	if doc == nil {
		t.Errorf("Should have not gotten nil document: err: %v", err)
	}
}

func TestServiceSaveGetDocumentErr(t *testing.T) {
	service, db := initService(t)
	defer db.DropTable(&did.PostgresDocument{})

	doc := did.BuildTestDocument()

	// Save document
	err := service.SaveDocument(doc)
	if err != nil {
		t.Errorf("Should have not gotten error saving document: err: %v", err)
	}

	// Get document with invalid DID
	doc, err = service.GetDocument("thisisnotadid")
	if err == nil {
		t.Errorf("Should have gotten error retrieving document")
	}
	if doc != nil {
		t.Errorf("Should have gotten nil document: err: %v", err)
	}

	// Get document with unknown DID
	doc, err = service.GetDocument("did:example:1234567")
	if err != nil {
		t.Errorf("Should have not gotten error retrieving document for unknown DID")
	}
	if doc != nil {
		t.Errorf("Should have gotten nil document: err: %v", err)
	}

	// Get document via a nil did
	doc, err = service.GetDocumentFromDID(nil)
	if err == nil {
		t.Errorf("Should have gotten error retrieving document from nil DID")
	}
	if doc != nil {
		t.Errorf("Should have gotten nil document: err: %v", err)
	}
}

func TestGenerateEthURIDID(t *testing.T) {
	service, db := initService(t)
	defer db.DropTable(&did.PostgresDocument{})

	newDID, err := service.GenerateEthURIDID()
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
	service, db := initService(t)
	defer db.DropTable(&did.PostgresDocument{})

	newDID, err := service.GenerateEthURIDID()
	if err != nil {
		t.Errorf("Should have not gotten error on DID generation")
	}
	firstPK := &did.DocPublicKey{
		ID:              *newDID,
		Type:            did.LDSuiteTypeSecp256k1Verification,
		Controller:      newDID,
		EthereumAddress: "0x5E4A048a9B8F5256a0D485e86E31e2c3F86523FB",
	}
	newDoc, err := service.InitializeNewDocument(newDID, firstPK)
	if err != nil {
		t.Errorf("Should not have gotten error generating new doc")
	}

	if newDoc.ID.String() != newDID.String() {
		t.Errorf("Should not have gotten same DID")
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

	bys, _ := json.Marshal(newDoc)
	t.Logf("%v", string(bys))
}
