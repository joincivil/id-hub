package idhubmain

import (
	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
)

func initDidPersister(db *gorm.DB) did.Persister {
	persister := did.NewPostgresPersister(db)
	db.AutoMigrate(did.PostgresDocument{})
	return persister
}

func initNodePersister(db *gorm.DB) *claimsstore.NodePGPersister {
	persister := claimsstore.NewNodePGPersisterWithDB(db)
	db.AutoMigrate(claimsstore.Node{})
	return persister
}

func initTreePersister(db *gorm.DB) *claimsstore.PGStore {
	nodePersister := initNodePersister(db)
	persister := claimsstore.NewPGStore(nodePersister)
	return persister
}

func initSignedClaimPersister(db *gorm.DB) *claimsstore.SignedClaimPGPersister {
	persister := claimsstore.NewSignedClaimPGPersister(db)
	db.AutoMigrate(claimsstore.SignedClaimPostgres{})
	return persister
}
