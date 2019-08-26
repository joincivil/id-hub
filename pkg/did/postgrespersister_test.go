package did

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

func BuildTestDocument() *Document {
	doc := &Document{}

	mainDID, _ := didlib.Parse(testDID)

	doc.ID = *mainDID
	doc.Context = DefaultDIDContextV1
	doc.Controller = mainDID

	// Public Keys
	pk1 := DocPublicKey{}
	pk1ID := fmt.Sprintf("%v#keys-1", testDID)
	d, _ := didlib.Parse(pk1ID)
	pk1.ID = d
	pk1.Type = LDSuiteTypeRsaSignature
	pk1.Controller = mainDID
	pk1.PublicKeyPem = "thisisapem"

	doc.PublicKeys = []DocPublicKey{pk1}

	// Service endpoints
	ep1 := DocService{}
	ep1ID := fmt.Sprintf("%v#vcr", testDID)
	d, _ = didlib.Parse(ep1ID)
	ep1.ID = *d
	ep1.Type = "CredentialRepositoryService"
	ep1.ServiceEndpoint = "https://repository.example.com/service/8377464"

	doc.Services = []DocService{ep1}

	// Authentication
	aw1 := DocAuthenicationWrapper{}
	aw1ID := fmt.Sprintf("%v#keys-1", testDID)
	d, _ = didlib.Parse(aw1ID)
	aw1.ID = d
	aw1.IDOnly = true

	aw2 := DocAuthenicationWrapper{}
	aw2ID := fmt.Sprintf("%v#keys-2", testDID)
	d, _ = didlib.Parse(aw2ID)
	aw2.ID = d
	aw2.Type = LDSuiteTypeRsaSignature
	aw2.Controller = mainDID
	aw2.PublicKeyPem = "thisisanotherpem"

	doc.Authentications = []DocAuthenicationWrapper{aw1, aw2}

	return doc
}

func TestSaveGetDocument(t *testing.T) {
	persister, err := setupTestTable()
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	// defer deleteTestTable(persister) // nolint: errcheck

	// Save a document
	testDoc := BuildTestDocument()
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
