package claims

import (
	"crypto/ecdsa"
	"encoding/hex"
	"time"

	ecrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

// Signer interface is for signing content claims
type Signer interface {
	Sign(claim *claimtypes.ContentCredential, creator string) error
}

// ECDSASigner implements the signer interface for a given private key
type ECDSASigner struct {
	privateKey *ecdsa.PrivateKey
}

// NewECDSASigner returns a new ecdsa signer
func NewECDSASigner(privKey *ecdsa.PrivateKey) *ECDSASigner {
	return &ECDSASigner{
		privateKey: privKey,
	}
}

// Sign takes a credential and a creator did and adds the proof
func (s ECDSASigner) Sign(claim *claimtypes.ContentCredential, creator string) error {
	canonical, err := CanonicalizeCredential(claim)
	if err != nil {
		return err
	}
	hash := ecrypto.Keccak256(canonical)
	sigBytes, err := ecrypto.Sign(hash, s.privateKey)
	if err != nil {
		return err
	}

	proofValue := hex.EncodeToString(sigBytes)
	proof := linkeddata.Proof{
		Type:       string(linkeddata.SuiteTypeSecp256k1Signature),
		Creator:    creator,
		Created:    time.Now(),
		ProofValue: proofValue,
	}
	proofSlice := []interface{}{proof}
	claim.Proof = proofSlice
	return nil
}
