package testinits

import (
	"math/big"

	"github.com/jinzhu/gorm"
	"github.com/joincivil/go-common/pkg/lock"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
)

// MakeService makes a new claims service
func MakeService(db *gorm.DB, didService *did.Service,
	signedClaimStore *claimsstore.SignedClaimPGPersister) (*claims.Service, *claims.RootService, error) {
	nodepersister := claimsstore.NewNodePGPersisterWithDB(db)
	treeStore := claimsstore.NewPGStore(nodepersister)
	rootCommitStore := claimsstore.NewRootCommitsPGPersister(db)
	dlock := lock.NewLocalDLock()
	committer := &claims.FakeRootCommitter{CurrentBlockNumber: big.NewInt(1)}
	rootService, _ := claims.NewRootService(treeStore, committer, rootCommitStore)
	claimService, err := claims.NewService(treeStore, signedClaimStore, didService, rootService, dlock)
	return claimService, rootService, err
}
