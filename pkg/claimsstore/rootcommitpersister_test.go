package claimsstore_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/testutils"
)

func setupRootCommitsConnection() (*gorm.DB, error) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&claimsstore.SignedClaimPostgres{}, &claimsstore.RootCommit{}).Error
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestSaveAndGet(t *testing.T) {
	db, err := setupRootCommitsConnection()
	if err != nil {
		t.Fatalf("couldn't set up db connection: %v", err)
	}
	persister := claimsstore.NewRootCommitsPGPersister(db)
	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	commit := &claimsstore.RootCommit{
		Root:             "someroothashorwhatevs",
		BlockNumber:      5,
		Prefix:           "roottree",
		TransactionHash:  "0xwhatevs",
		ContractAddress:  "0xcontractaddress",
		CommitterAddress: "0xsomebodiesaddress",
	}
	err = persister.Save(commit)
	if err != nil {
		t.Errorf("should not error when saving: %v", err)
	}
	commit2, err := persister.Get(commit.Root)
	if err != nil {
		t.Errorf("should not error when getting: %v", err)
	}
	if commit.TransactionHash != commit2.TransactionHash {
		t.Errorf("retrieved commit should match saved commit")
	}
}

func TestGetLatest(t *testing.T) {
	db, err := setupRootCommitsConnection()
	if err != nil {
		t.Fatalf("couldn't set up db connection: %v", err)
	}
	persister := claimsstore.NewRootCommitsPGPersister(db)
	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()
	commit := &claimsstore.RootCommit{
		Root:             "someroothashorwhatevs",
		BlockNumber:      5,
		Prefix:           "roottree",
		TransactionHash:  "0xwhatevs",
		ContractAddress:  "0xcontractaddres",
		CommitterAddress: "0xsomebodiesaddress",
	}
	commit2 := &claimsstore.RootCommit{
		Root:             "someroothashorwhatevs2",
		BlockNumber:      6,
		Prefix:           "roottree",
		TransactionHash:  "0xwhatevs2",
		ContractAddress:  "0xcontractaddres",
		CommitterAddress: "0xsomebodiesaddress",
	}
	commit3 := &claimsstore.RootCommit{
		Root:             "someroothashorwhatevs3",
		BlockNumber:      7,
		Prefix:           "roottree",
		TransactionHash:  "0xwhatevs3",
		ContractAddress:  "0xcontractaddres",
		CommitterAddress: "0xsomebodiesaddress",
	}
	err = persister.Save(commit)
	if err != nil {
		t.Errorf("should not error when saving: %v", err)
	}
	err = persister.Save(commit2)
	if err != nil {
		t.Errorf("should not error when saving: %v", err)
	}
	err = persister.Save(commit3)
	if err != nil {
		t.Errorf("should not error when saving: %v", err)
	}
	commit4, err := persister.GetLatest()
	if err != nil {
		t.Errorf("should not error when getting latest: %v", err)
	}
	if commit3.Root != commit4.Root {
		t.Errorf("should have returned the last root saved")
	}

}
