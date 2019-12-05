package claimtypes

import (
	"bytes"
	"encoding/json"
	"errors"

	didhub "github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"

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

// FindLinkedDataProof returns the first the linked data proof in the proof slice
func FindLinkedDataProof(proofs interface{}) (*linkeddata.Proof, error) {
	var err error
	// If this is an interface slice
	proofList, ok := proofs.([]interface{})
	if ok {
		var p *linkeddata.Proof
		for _, v := range proofList {
			p, err = ConvertToLinkedDataProof(v)
			if err != nil {
				continue
			}
			return p, nil
		}
		return nil, errors.New("proofs array didn't contain a valid linked data proof")
	}

	// If it is not an interface slice
	return ConvertToLinkedDataProof(proofs)
}

// ConvertToLinkedDataProof returns a linkeddata.Proof from the proof interface{} value
func ConvertToLinkedDataProof(proof interface{}) (*linkeddata.Proof, error) {
	switch val := proof.(type) {
	case *linkeddata.Proof:
		return val, nil

	case linkeddata.Proof:
		return &val, nil

	case map[string]interface{}:
		t, ok := val["type"]
		if ok && t == string(linkeddata.SuiteTypeSecp256k1Signature) {
			ld := &linkeddata.Proof{}
			js, err := json.Marshal(val)

			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(js, ld)
			if err != nil {
				return nil, err
			}
			return ld, nil
		}
	}

	return nil, errors.New("proof was not a valid linked data proof")
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
