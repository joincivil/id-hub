package ethuri

import (
	"fmt"
	"testing"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
	"github.com/joincivil/id-hub/pkg/testutils"
	didlib "github.com/ockam-network/did"
)

const (
	testDID = "did:ethuri:fbaf6bb3-2a82-4173-b31a-160a143c931c"
)

func setupDBConnection() (*PostgresPersister, error) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&PostgresDocument{}).Error
	if err != nil {
		return nil, err
	}

	return NewPostgresPersister(db), nil
}

func setupTestTable() (*PostgresPersister, error) {
	persister, err := setupDBConnection()
	if err != nil {
		return persister, fmt.Errorf("Error connecting to DB: %v", err)
	}
	return persister, nil
}

func deleteTestTable(persister *PostgresPersister) error {
	return persister.db.DropTable(&PostgresDocument{}).Error
}

func TestSaveGetDocument(t *testing.T) {
	persister, err := setupTestTable()
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	// defer deleteTestTable(persister) // nolint: errcheck

	// Save a document
	testDoc := testutils.BuildTestDocument()
	err = persister.SaveDocument(testDoc)
	if err != nil {
		t.Errorf("Should have saved the document: err: %v", err)
	}

	// Get the document
	d, _ := didlib.Parse(testDID)
	doc, err := persister.GetDocument(d)
	if err != nil {
		t.Fatalf("Should have retrieved a document: err: %v", err)
	}
	if doc.ID.String() != testDID {
		t.Errorf("Should have gotten back the same ID")
	}
}

func TestSaveGetDocumentErr(t *testing.T) {
	persister, err := setupTestTable()
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister) // nolint: errcheck

	// Get an unknown did
	d, err := didlib.Parse("did:example:123456")
	if err != nil {
		t.Errorf("Should have not returned an error parsing an invalid did")
	}
	if d == nil {
		t.Errorf("Should have not returned a nil did")
	}

	// Get an unknown document
	doc, err := persister.GetDocument(d)
	if err == nil {
		t.Errorf("Should have returned an error getting unknown did")
	}
	if err != cpersist.ErrPersisterNoResults {
		t.Errorf("Should have returned ErrPersisterNoResults")
	}
	if doc != nil {
		t.Errorf("Should have returned a nil document")
	}

	// Get an invalid did
	d, err = didlib.Parse("invaliddid")
	if err == nil {
		t.Errorf("Should have returned an error parsing an invalid did")
	}
	if d != nil {
		t.Errorf("Should have returned a nil did")
	}

	// Get a nil document
	doc, err = persister.GetDocument(d)
	if err == nil {
		t.Errorf("Should have returned an error getting invalid did")
	}
	if doc != nil {
		t.Errorf("Should have returned a nil document")
	}
}
