package claimsstore

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type Claim struct {
	gorm.Model
	DID       string
	ClaimData string
	// ClaimJSON postgres.Jsonb
	ClaimKey string `gorm:"unique_index;not null;"`
	NodeType string
}

// TableName sets the name of the corresponding table in the db
func (Claim) TableName() string {
	return "claims"
}

type ClaimPGPersister struct {
	DB *gorm.DB
}

// NewClaimPGPersister return a new persister
func NewClaimPGPersister(host string, port int, user string, password string, dbname string) (*ClaimPGPersister, error) {
	gormPGPersister := &ClaimPGPersister{}
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := gorm.Open("postgres", psqlInfo)
	if err != nil {
		return gormPGPersister, errors.Wrap(err, "error connecting to gorm")
	}
	gormPGPersister.DB.DB().SetMaxOpenConns(maxOpenConns)
	gormPGPersister.DB.DB().SetMaxIdleConns(maxIdleConns)
	gormPGPersister.DB.DB().SetConnMaxLifetime(connMaxLifetime)
	return gormPGPersister, nil
}

// NewClaimPGPersister uses an existing gorm.DB struct to create a new GormPGPersister.
// This is useful if we want to reuse existing connections
func NewClaimPGPersisterWithDB(db *gorm.DB) (*ClaimPGPersister, error) {
	gormPGPersister := &ClaimPGPersister{}
	gormPGPersister.DB = db
	return gormPGPersister, nil
}

func (c *ClaimPGPersister) Add(claim *Claim) error {

}
