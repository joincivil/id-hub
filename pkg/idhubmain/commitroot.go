package idhubmain

import (
	log "github.com/golang/glog"
)

// RunCommitRoot whatever the current root is for the root_merkletree is saved to the smart contract
func RunCommitRoot() error {
	config := populateConfig()

	db, err := initGorm(config)
	if err != nil {
		log.Fatalf("error initializing gorm: %v", err)
	}

	persister := initRootClaimPersister(db)

	ethHelper, err := initETHHelper(config)
	if err != nil {
		log.Fatalf("error initializing eth helper: %v", err)
	}

	treeStore := initTreePersister(db)

	rootService, err := initRootService(config, ethHelper, treeStore, persister)
	if err != nil {
		log.Fatalf("error initializing root service: %v", err)
	}

	return rootService.CommitRoot()
}
