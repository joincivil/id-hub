package claims_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/go-common/pkg/generated/contract"
	"github.com/joincivil/id-hub/pkg/claims"
)

func deployContract(helper *eth.Helper) (common.Address, *types.Transaction, *contract.RootCommitsContract, error) {
	return contract.DeployRootCommitsContract(helper.Auth, helper.Blockchain)
}

func TestCommitRoot(t *testing.T) {
	ethHelper, err := eth.NewSimulatedBackendHelper()
	blockchain := ethHelper.Blockchain.(*backends.SimulatedBackend)
	if err != nil {
		t.Fatalf("error constructing blockchain helper: err: %v", err)
	}

	contractAddress, _, _, err := deployContract(ethHelper)
	if err != nil {
		t.Fatalf("error deploying root commit contract: err: %v", err)
	}
	blockchain.Commit()

	rootCommitter, err := claims.NewRootCommitter(ethHelper, blockchain, contractAddress.String())
	if err != nil {
		t.Fatalf("error creating root commiter: %v", err)
	}
	c := make(chan *claims.ProgressUpdate)
	root := [32]byte{
		0x0, 0x3, 0x0, 0x3, 0x0, 0x3, 0x0, 0x3,
		0x0, 0x3, 0x0, 0x3, 0x0, 0x3, 0x0, 0x3,
		0x0, 0x3, 0x0, 0x3, 0x0, 0x3, 0x0, 0x3,
		0x0, 0x3, 0x0, 0x3, 0x0, 0x3, 0x0, 0x3,
	}
	go rootCommitter.CommitRoot(root, c)
	var result *claims.ProgressUpdate
	for res := range c {
		if res.Status == claims.Started {
			blockchain.Commit()
		} else {
			result = res
		}
	}

	if result.Err != nil {
		t.Fatalf("error commiting root and getting receipt: %v", result.Err)
	}

	if result.Result.BlockNumber.Int64() != 2 {
		t.Errorf("wrong block number")
	}
}
