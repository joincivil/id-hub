package did

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/joincivil/id-hub/pkg/testutils"
)

const (
	testKeyFilename = "unit.test.pub.key"
	testKeyVal      = "046539bd140ab14032735641692cbc3e7b52ef9e367887f4f2fd53942c870a5279c8639a511d9965c56c13fc7b00e636ecf0ea77237dd3e363a31ce95a06e58080"
)

func writeTestKeyFile() error {
	return ioutil.WriteFile(testKeyFilename, []byte(testKeyVal), 0777)
}

func deleteTestKeyFile() error {
	return os.Remove(testKeyFilename)
}

func persister(t *testing.T) *PostgresPersister {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		t.Fatal("Should have gotten test gorm")
	}
	db.AutoMigrate(&PostgresDocument{})
	p := NewPostgresPersister(db)
	return p
}

func TestGenerateDIDCli(t *testing.T) {
	err := writeTestKeyFile()
	if err != nil {
		t.Fatalf("Should have written the test key file")
	}

	defer func() {
		err := deleteTestKeyFile()
		if err != nil {
			t.Fatalf("Should have written the test key file")
		}
	}()

	// Happy case
	_, err = GenerateDIDCli(LDSuiteTypeSecp256k1Verification, testKeyFilename, nil)
	if err != nil {
		t.Fatalf("Should have not gotten error for did cli: %v", err)
	}

	// Wrong signature type
	_, err = GenerateDIDCli(LDSuiteTypeRsaSignature, testKeyFilename, nil)
	if err == nil {
		t.Fatalf("Should have gotten error for did cli")
	}

	// Wrong key file name
	_, err = GenerateDIDCli(LDSuiteTypeRsaSignature, "wrong.key.file", nil)
	if err == nil {
		t.Fatalf("Should have gotten error for did cli")
	}

	// Test the DID store to postgres
	p := persister(t)

	// postgrespersister_test
	defer deleteTestTable(p) // nolint: errcheck

	doc, err := GenerateDIDCli(LDSuiteTypeSecp256k1Verification, testKeyFilename, p)
	if err != nil {
		t.Fatalf("Should have not gotten error for did cli gen")
	}

	if doc == nil {
		t.Fatalf("Should not have returned a nil doc")
	}

	fromdb, err := p.GetDocument(&doc.ID)
	if err != nil {
		t.Errorf("Should not have gotten error retrieving did")
	}

	if fromdb.ID.String() != doc.ID.String() {
		t.Errorf("Should have matched DIDs")
	}
}
