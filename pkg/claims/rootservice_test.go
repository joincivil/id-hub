package claims_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	iden3db "github.com/iden3/go-iden3-core/db"
	"github.com/iden3/go-iden3-core/merkletree"
	isrv "github.com/iden3/go-iden3-core/services/claimsrv"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/testutils"
	didlib "github.com/ockam-network/did"
)

type fakeRootCommitter struct{}

func (r *fakeRootCommitter) CommitRoot(root [32]byte, c chan<- *claims.ProgressUpdate) {
	defer close(c)
	c <- &claims.ProgressUpdate{Status: claims.Done, Result: &ethTypes.Receipt{
		BlockNumber:     big.NewInt(2),
		TxHash:          common.HexToHash("0x368782c63319f79c83cb937fefe1f0268c6fd098930e1d590a45ad233bcace37"),
		ContractAddress: common.HexToAddress("0x6BBDd7B1a289C5bE8fAa29Cb1c0be66cb2582060"),
	}, Err: nil}
}

func makeRootService(db *gorm.DB) (*claims.RootService, *claimsstore.RootCommitsPGPersister, iden3db.Storage, error) {
	nodepersister := claimsstore.NewNodePGPersisterWithDB(db)
	rootpersister := claimsstore.NewRootCommitsPGPersister(db)
	treeStore := claimsstore.NewPGStore(nodepersister)
	committer := &fakeRootCommitter{}
	rootService, err := claims.NewRootService(treeStore, committer, rootpersister)
	return rootService, rootpersister, treeStore, err
}

func addNewRootClaim(mt *merkletree.MerkleTree, userDid *didlib.DID) error {
	root := merkletree.Hash(merkletree.ElemBytes{
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0c})

	claimSetRootKey, err := claimtypes.NewClaimSetRootKeyDID(userDid, &root)
	if err != nil {
		return err
	}
	// get next version of the claim
	version, err := isrv.GetNextVersion(mt, claimSetRootKey.Entry().HIndex())
	if err != nil {
		return err
	}
	claimSetRootKey.Version = version
	err = mt.Add(claimSetRootKey.Entry())
	if err != nil {
		return err
	}
	return nil
}

func TestRootServiceCommitRoot(t *testing.T) {
	db, err := setupConnection()
	if err != nil {
		t.Errorf("error setting up the db: %v", err)
	}

	cleaner := testutils.DeleteCreatedEntities(db)
	defer cleaner()

	rootService, persister, treeStore, err := makeRootService(db)
	if err != nil {
		t.Errorf("error creating root service: %v", err)
	}

	rootStore := treeStore.WithPrefix(claimsstore.PrefixRootMerkleTree)

	rootMt, err := merkletree.NewMerkleTree(rootStore, 150)
	if err != nil {
		t.Errorf("error creating root merkletree")
	}

	userDIDs := "did:ethuri:86ce6c71-27e6-4e0d-83dd-b60fe4d7785c"
	userDID, err := didlib.Parse(userDIDs)
	if err != nil {
		t.Errorf("error parsing did: %v", err)
	}

	err = addNewRootClaim(rootMt, userDID)
	if err != nil {
		t.Errorf("error adding a claim to the merkle tree: %v", err)
	}

	root := rootMt.RootKey()

	err = rootService.CommitRoot()
	if err != nil {
		t.Errorf("error committing the root: %v", err)
	}

	rootCommit, err := persister.Get(root.Hex())

	if err != nil {
		t.Errorf("error fetching root commit: %v", err)
	}
	fmt.Printf("\n\n%v\n\n", rootCommit)
	if rootCommit.BlockNumber != 2 {
		t.Errorf("block number did not match expected")
	}

	if rootCommit.TransactionHash != "0x368782c63319f79c83cb937fefe1f0268c6fd098930e1d590a45ad233bcace37" {
		t.Errorf("transaction hash did not match expected")
	}

}
