package claims

import (
	"context"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum"
	ethCommon "github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/go-common/pkg/generated/contract"
	"github.com/joincivil/go-common/pkg/jobs"
)

const (
	// Done signifies indexing is finished or errored
	Done = "done"
	// Started signifies indexing has begun
	Started = "started"
)

// RootCommitterInterface specifies the interface of the struct that interacts with the blockchain
type RootCommitterInterface interface {
	CommitRoot(root [32]byte, c chan<- *ProgressUpdate)
}

// RootCommitter performs the transaction that commits the root to the blockchain and awaits completion
type RootCommitter struct {
	ethHelper         *eth.Helper
	txListener        *eth.TxListener
	rootContract      *contract.RootCommitsContract
	transactionReader ethereum.TransactionReader
}

// ProgressUpdate format for passing status of the transaction to the main routine
type ProgressUpdate struct {
	Status string
	Result *ethTypes.Receipt
	Err    error
}

// NewRootCommitter constructs a new root committer
func NewRootCommitter(ethHelper *eth.Helper, transactionReader ethereum.TransactionReader, address string) (*RootCommitter, error) {
	txListener := eth.NewTxListenerWithWaitPeriod(transactionReader, jobs.NewInMemoryJobService(), 2*time.Minute)
	contractAddress := ethCommon.HexToAddress(address)
	if contractAddress == ethCommon.HexToAddress("") {
		return nil, errors.New("must have a valid address for the root commit contract")
	}

	rootContract, err := contract.NewRootCommitsContract(contractAddress, ethHelper.Blockchain)
	if err != nil {
		return nil, err
	}
	return &RootCommitter{
		ethHelper:         ethHelper,
		txListener:        txListener,
		rootContract:      rootContract,
		transactionReader: transactionReader,
	}, nil
}

// CommitRoot given a root performs the transaction to add it to the contract
func (r *RootCommitter) CommitRoot(root [32]byte, c chan<- *ProgressUpdate) {
	defer close(c)
	tx, err := r.rootContract.SetRoot(r.ethHelper.Transact(), root)
	if err != nil {
		c <- &ProgressUpdate{Status: Done, Result: &ethTypes.Receipt{}, Err: err}
		return
	}
	txHash := tx.Hash()
	c <- &ProgressUpdate{Status: Started, Result: &ethTypes.Receipt{}, Err: nil}
	sub, err := r.txListener.StartListener(txHash.String())
	if err != nil {
		c <- &ProgressUpdate{Status: Done, Result: &ethTypes.Receipt{}, Err: err}
		return
	}
	for range sub.Updates {
		// wait for channel to close
	}
	ctx := context.Background()
	receipt, err := r.transactionReader.TransactionReceipt(ctx, txHash)
	if err != nil {
		c <- &ProgressUpdate{Status: Done, Result: &ethTypes.Receipt{}, Err: err}
		return
	}
	c <- &ProgressUpdate{Status: Done, Result: receipt, Err: nil}
}
