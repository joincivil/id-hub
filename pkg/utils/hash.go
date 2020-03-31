package utils

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiformats/go-multihash"
)

// CreateMultihash creates a multihash given bytes
func CreateMultihash(bytes []byte) ([]byte, error) {
	hash := crypto.Keccak256(bytes)
	return multihash.EncodeName(hash, "keccak-256")
}

// MultiHashString create a multihash hex from a string
func MultiHashString(str string) (string, error) {
	mHash, err := CreateMultihash([]byte(str))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mHash), nil
}
