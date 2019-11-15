package claimtypes

import (
	"bytes"

	didhub "github.com/joincivil/id-hub/pkg/did"

	"github.com/iden3/go-iden3-core/merkletree"
	"github.com/iden3/go-iden3-crypto/poseidon"
	didlib "github.com/ockam-network/did"
)

// HashDID uses poseidon to hash the did to create a hash that fits in the merkletree's
// 31 byte limit
func HashDID(did *didlib.DID) ([]byte, error) {
	bigInts, err := poseidon.HashBytes([]byte(didhub.MethodIDOnly(did)))
	if err != nil {
		return nil, err
	}
	return merkletree.BigIntToHash(bigInts).Bytes(), nil
}

// Concat is  a internal method from iden3 db that seemed necessary to implement the interface
func Concat(vs ...[]byte) []byte {
	var b bytes.Buffer
	for _, v := range vs {
		b.Write(v)
	}
	return b.Bytes()
}

// these are methods copied from iden3 that they made private but are very useful

// copyToElemBytes copies the src slice forwards to e, ending at -start of
// e.  This function will panic if src doesn't fit into len(e)-start.
func copyToElemBytes(e *merkletree.ElemBytes, start int, src []byte) {
	copy(e[merkletree.ElemBytesLen-start-len(src):], src)
}

// copyFromElemBytes copies from e to dst, ending at -start of e and going
// forwards.  This function will panic if len(e)-start is smaller than
// len(dst).
func copyFromElemBytes(dst []byte, start int, e *merkletree.ElemBytes) {
	copy(dst, e[merkletree.ElemBytesLen-start-len(dst):])
}
