package claimsstore

import (
	"github.com/jinzhu/gorm"
)

// RootCommit stores information about a root saved to the contract
type RootCommit struct {
	gorm.Model
	Root             string `gorm:"primary_key"`
	BlockNumber      int64
	Prefix           string
	TransactionHash  string
	ContractAddress  string
	CommitterAddress string
}

// TableName sets the table name for signed claims
func (RootCommit) TableName() string {
	return "root_commits"
}

// RootCommitsPGPersister persister model for root commits
type RootCommitsPGPersister struct {
	db *gorm.DB
}

// NewRootCommitsPGPersister returns a new RootCommitsPGPersister
func NewRootCommitsPGPersister(db *gorm.DB) *RootCommitsPGPersister {
	return &RootCommitsPGPersister{
		db: db,
	}
}

// Save saves a new root commit to the db
func (p *RootCommitsPGPersister) Save(root *RootCommit) error {
	if err := p.db.Create(root).Error; err != nil {
		return err
	}
	return nil
}

// Get returns the information about a root commit given a root hash
func (p *RootCommitsPGPersister) Get(rootHash string) (*RootCommit, error) {
	rootCommit := &RootCommit{}
	if err := p.db.Where("root = ?", rootHash).First(rootCommit).Error; err != nil {
		return rootCommit, err
	}
	return rootCommit, nil
}

// GetLatest returns the most recent root committed to the tree
func (p *RootCommitsPGPersister) GetLatest() (*RootCommit, error) {
	rootCommit := &RootCommit{}
	stmt := p.db.Raw("SELECT * FROM root_commits WHERE block_number = (SELECT MAX (block_number) FROM root_commits)")
	if err := stmt.Scan(rootCommit).Error; err != nil {
		return rootCommit, err
	}
	return rootCommit, nil
}
