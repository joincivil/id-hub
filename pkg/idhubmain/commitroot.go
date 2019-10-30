package idhubmain

import (
	"github.com/ethereum/go-ethereum"
	log "github.com/golang/glog"
	"github.com/joincivil/id-hub/pkg/claims"
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

	rootCommitter, err := claims.NewRootCommitter(ethHelper, ethHelper.Blockchain.(ethereum.TransactionReader), config.RootCommitsAddress)
	if err != nil {
		log.Fatalf("error initializing root committer: %v", err)
	}

	treeStore := initTreeStore(db)

	rootService, err := claims.NewRootService(treeStore, rootCommitter, persister)
	if err != nil {
		log.Fatalf("error initializing root service: %v", err)
	}

	return rootService.CommitRoot()
}
