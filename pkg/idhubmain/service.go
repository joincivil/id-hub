package idhubmain

import (
	"github.com/ethereum/go-ethereum"
	log "github.com/golang/glog"
	"github.com/iden3/go-iden3-core/db"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/go-common/pkg/lock"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/utils"
)

func initDidService(resolvers []did.Resolver) *did.Service {
	return did.NewService(resolvers)
}

func initClaimsService(treeStore *claimsstore.PGStore, signedClaimStore *claimsstore.SignedClaimPGPersister,
	didService *did.Service, rootService *claims.RootService, dlock lock.DLock) (*claims.Service, error) {
	return claims.NewService(treeStore, signedClaimStore, didService, rootService, dlock)
}

func initRootService(config *utils.IDHubConfig, ethHelper *eth.Helper,
	treeStore db.Storage, persister *claimsstore.RootCommitsPGPersister) (*claims.RootService, error) {
	if config.RootCommitsAddress == "" {
		log.Errorf("No root commits address set, disabling root commits access")
		return nil, nil
	}

	rootCommitter, err := claims.NewRootCommitter(
		ethHelper,
		ethHelper.Blockchain.(ethereum.TransactionReader),
		config.RootCommitsAddress,
	)
	if err != nil {
		return nil, err
	}
	return claims.NewRootService(treeStore, rootCommitter, persister)
}
