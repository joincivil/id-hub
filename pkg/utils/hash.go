package utils

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiformats/go-multihash"
)

// CreateMultihash creates a multihash given bytes
func CreateMultihash(bytes []byte) ([]byte, error) {
	hash := crypto.Keccak256(bytes)
	return multihash.EncodeName(hash, "keccak-256")
}
