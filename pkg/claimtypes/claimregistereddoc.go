package claimtypes

import (
	"encoding/binary"

	"github.com/iden3/go-iden3-core/core"
	"github.com/iden3/go-iden3-core/merkletree"
	didlib "github.com/ockam-network/did"
)

// ClaimTypeRegisteredDocument is the type indicator for the registered document claim
var ClaimTypeRegisteredDocument = core.NewClaimTypeNum(11)

const (
	// ContentCredentialDocType is an enum value that differentiates content credentials from other document types that may be registered later
	ContentCredentialDocType uint32 = iota
)

// ClaimRegisteredDocument is a claim type for registering other claims like signed claims in the merkle tree
type ClaimRegisteredDocument struct {
	Version     uint32
	ContentHash [34]byte
	DID         [32]byte
	DocType     uint32
}

// NewClaimRegisteredDocument creates a new ClaimRegisteredDocument from a did a contentHash and a type
func NewClaimRegisteredDocument(ch [34]byte, did *didlib.DID, dt uint32) (*ClaimRegisteredDocument, error) {
	didbytes, err := HashDID(did)
	if err != nil {
		return nil, err
	}
	didbytes32 := [32]byte{}
	copy(didbytes32[:], didbytes[:])
	return &ClaimRegisteredDocument{
		Version:     0,
		DID:         didbytes32,
		ContentHash: ch,
		DocType:     dt,
	}, nil
}

// NewClaimRegisteredDocumentFromEntry turns a merkletree Entry into a ClaimRegisteredDocument
func NewClaimRegisteredDocumentFromEntry(e *merkletree.Entry) *ClaimRegisteredDocument {
	c := &ClaimRegisteredDocument{}
	_, c.Version = core.GetClaimTypeVersionFromData(&e.Data)
	var docType [32 / 8]byte
	copyFromElemBytes(docType[:], 4, &e.Data[0])
	c.DocType = binary.BigEndian.Uint32(docType[:])

	copyFromElemBytes(c.DID[:], 0, &e.Data[1])

	hashBeginning := [3]byte{}
	hashRest := [31]byte{}
	copyFromElemBytes(hashBeginning[:], 4, &e.Data[0])
	copyFromElemBytes(hashRest[:], 0, &e.Data[2])
	contentHash := Concat(hashBeginning[:], hashRest[:])
	copy(c.ContentHash[:], contentHash)
	return c
}

// Entry converts the ClaimRegisteredDocument into a merkletree entry
func (c ClaimRegisteredDocument) Entry() *merkletree.Entry {
	e := &merkletree.Entry{}
	core.SetClaimTypeVersion(e, c.Type(), c.Version)
	var docType [4]byte
	binary.BigEndian.PutUint32(docType[:], c.DocType)
	copyToElemBytes(&e.Data[0], 0, docType[:])
	copyToElemBytes(&e.Data[0], 4, c.ContentHash[0:3])
	copyToElemBytes(&e.Data[2], 0, c.ContentHash[3:])
	copyToElemBytes(&e.Data[1], 0, c.DID[:])
	return e
}

// Type returns the type of the claim
func (c *ClaimRegisteredDocument) Type() core.ClaimType {
	return *ClaimTypeRegisteredDocument
}
