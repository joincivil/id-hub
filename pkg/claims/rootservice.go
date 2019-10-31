package claims

import (
	"github.com/iden3/go-iden3-core/db"
	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/joincivil/id-hub/pkg/claimsstore"
)

// RootService coordinates publishing the root to the blockchain and saving the result to pg
type RootService struct {
	treeStore db.Storage
	committer RootCommitterInterface
	persister *claimsstore.RootCommitsPGPersister
}

// NewRootService constructs a new root service
func NewRootService(treeStore db.Storage, committer RootCommitterInterface, persister *claimsstore.RootCommitsPGPersister) (*RootService, error) {
	return &RootService{
		treeStore: treeStore,
		committer: committer,
		persister: persister,
	}, nil
}

// CommitRoot commits the current root of the root tree to the contract and saves the blocknumber and transaction in pg
func (s *RootService) CommitRoot() error {
	rootStore := s.treeStore.WithPrefix(claimsstore.PrefixRootMerkleTree)

	rootMt, err := merkletree.NewMerkleTree(rootStore, 150)
	if err != nil {
		return err
	}

	var root [32]byte
	rootSlice := rootMt.RootKey()
	copy(root[:], rootSlice.Bytes()[:32])
	c := make(chan *ProgressUpdate)
	go s.committer.CommitRoot(root, c)
	var result *ProgressUpdate
	for res := range c {
		if res.Status == Done {
			result = res
		}
	}
	if result.Err != nil {
		return result.Err
	}

	rootCommit := &claimsstore.RootCommit{
		Root:            rootSlice.Hex(),
		BlockNumber:     result.Result.BlockNumber.Int64(),
		Prefix:          string(claimsstore.PrefixRootMerkleTree),
		ContractAddress: result.Result.ContractAddress.String(),
		TransactionHash: result.Result.TxHash.String(),
	}

	return s.persister.Save(rootCommit)
}
