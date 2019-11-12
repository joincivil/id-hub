package idhubmain

import (
	"github.com/ethereum/go-ethereum"
	"github.com/iden3/go-iden3-core/db"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/utils"
)

func initDidService(persister did.Persister) *did.Service {
	return did.NewService(persister)
}

func initClaimsService(treeStore *claimsstore.PGStore, signedClaimStore *claimsstore.SignedClaimPGPersister,
	didService *did.Service, rootService *claims.RootService) (*claims.Service, error) {
	return claims.NewService(treeStore, signedClaimStore, didService, rootService)
}

func initRootService(config *utils.IDHubConfig, ethHelper *eth.Helper,
	treeStore db.Storage, persister *claimsstore.RootCommitsPGPersister) (*claims.RootService, error) {
	rootCommitter, err := claims.NewRootCommitter(ethHelper, ethHelper.Blockchain.(ethereum.TransactionReader), config.RootCommitsAddress)
	if err != nil {
		return nil, err
	}
	return claims.NewRootService(treeStore, rootCommitter, persister)
}
