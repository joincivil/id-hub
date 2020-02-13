package idhubmain

import (
	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/hedgehog"
)

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

func initRootClaimPersister(db *gorm.DB) *claimsstore.RootCommitsPGPersister {
	persister := claimsstore.NewRootCommitsPGPersister(db)
	db.AutoMigrate(
		claimsstore.RootCommit{},
	)
	return persister
}

func initHedgehog(db *gorm.DB) {
	db.AutoMigrate(
		hedgehog.DataVaultItem{},
	)
}
