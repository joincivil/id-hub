package claimsstore

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	didlib "github.com/ockam-network/did"
)

var (
	// PrefixRootMerkleTree prefix value for root tree
	PrefixRootMerkleTree = []byte("root_merkletree")
	// ErrTooLongDIDMethod did to binary is a bit fragile for now, fails when method is more than 15 bytes
	ErrTooLongDIDMethod = errors.New("method string is too long to fit in merkletree elembytes")
	// ErrWrongSizByteSliceDID if the slice is the wrong size can't convert it back to a did
	ErrWrongSizByteSliceDID = errors.New("binaryToDID expects a byte slice of length 32")
	singleEmptyByte         = []byte{32}
)

// DIDToBinary converts did into []bytes with 16 bytes as the uuid and 16 bytes as method
func DIDToBinary(did *didlib.DID) ([]byte, error) {
	uid, err := uuid.Parse(did.ID)
	if err != nil {
		return nil, err
	}
	method := []byte(did.Method)
	lenMethod := len(method)
	if lenMethod > 15 { // cant be longer then 14 need to bytes for store length of method
		return nil, ErrTooLongDIDMethod
	}
	lenMethodB := []byte(fmt.Sprintf("%2d", lenMethod))

	buid, err := uid.MarshalBinary()
	if err != nil {
		return nil, err
	}
	extraSpace := make([]byte, 14-lenMethod)

	return Concat(lenMethodB, method, extraSpace, buid), nil
}

// BinaryToDID converts a []byte to a did if it matches the 32 byte format
func BinaryToDID(b []byte) (*didlib.DID, error) {
	if len(b) != 32 {
		return nil, ErrWrongSizByteSliceDID
	}
	did := &didlib.DID{}
	methodLenb := b[:2]
	if bytes.Equal(b[:1], singleEmptyByte) {
		methodLenb = b[1:2]
	}
	methodLen, err := strconv.Atoi(string(methodLenb))
	if err != nil {
		return did, err
	}
	method := string(b[2 : 2+methodLen])
	uid, err := uuid.FromBytes(b[16:])
	if err != nil {
		return did, err
	}
	did.ID = uid.String()
	did.Method = method
	return did, nil
}

// Concat is  a internal method from iden3 db that seemed necessary to implement the interface
func Concat(vs ...[]byte) []byte {
	var b bytes.Buffer
	for _, v := range vs {
		b.Write(v)
	}
	return b.Bytes()
}
