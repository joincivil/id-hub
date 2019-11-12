package claimsstore

import (
	"bytes"
	"errors"
)

var (
	// PrefixRootMerkleTree prefix value for root tree
	PrefixRootMerkleTree = []byte("root_merkletree")
	// ErrTooLongDIDMethod did to binary is a bit fragile for now, fails when method is more than 15 bytes
	ErrTooLongDIDMethod = errors.New("method string is too long to fit in merkletree elembytes")
	// ErrWrongSizByteSliceDID if the slice is the wrong size can't convert it back to a did
	ErrWrongSizByteSliceDID = errors.New("binaryToDID expects a byte slice of length 32")
)

// Concat is  a internal method from iden3 db that seemed necessary to implement the interface
func Concat(vs ...[]byte) []byte {
	var b bytes.Buffer
	for _, v := range vs {
		b.Write(v)
	}
	return b.Bytes()
}
