package claims

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	didlib "github.com/ockam-network/did"
)

// AddProof takes a content cred a did and a pk and adds a proof to it
func AddProof(cred *claimtypes.ContentCredential, signerDID *didlib.DID, pk *ecdsa.PrivateKey) error {
	canonical, err := CanonicalizeCredential(cred)
	if err != nil {
		return err
	}
	hash := crypto.Keccak256(canonical)
	sigBytes, err := crypto.Sign(hash, pk)
	if err != nil {
		return err
	}

	proofValue := hex.EncodeToString(sigBytes)
	proofs := make([]interface{}, 0, 1)
	ld := linkeddata.Proof{
		Type:       string(linkeddata.SuiteTypeSecp256k1Signature),
		Creator:    signerDID.String(),
		Created:    time.Now(),
		ProofValue: proofValue,
	}
	proofs = append(proofs, ld)
	cred.Proof = proofs
	return nil
}

// FakeRootCommitter fakes the blockchain part of the committing roots for testing
type FakeRootCommitter struct{}

// CommitRoot fakely commits the root
func (r *FakeRootCommitter) CommitRoot(root [32]byte,
	c chan<- *ProgressUpdate) {
	defer close(c)
	c <- &ProgressUpdate{Status: Done, Result: &ethTypes.Receipt{
		BlockNumber:     big.NewInt(2),
		TxHash:          common.HexToHash("0x368782c63319f79c83cb937fefe1f0268c6fd098930e1d590a45ad233bcace37"),
		ContractAddress: common.HexToAddress("0x6BBDd7B1a289C5bE8fAa29Cb1c0be66cb2582060"),
	}, Err: nil}
}

// GetAccount returns an account that could have been the one used for testing
func (r *FakeRootCommitter) GetAccount() common.Address {
	return common.HexToAddress("0x9320352C9931267C003ED2a9E33f089e87a0F0EF")
}
