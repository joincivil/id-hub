package did_test

import (
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
