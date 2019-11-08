package claimtypes

import (
	"encoding/binary"
	"errors"

	"github.com/iden3/go-iden3-core/core"
	"github.com/iden3/go-iden3-core/merkletree"
	crconstants "github.com/iden3/go-iden3-crypto/constants"
	crutils "github.com/iden3/go-iden3-crypto/utils"
	didlib "github.com/ockam-network/did"
)

// ClaimTypeSetRootKeyDID this claimtype marks a new root key for a did tree
var ClaimTypeSetRootKeyDID = core.NewClaimTypeNum(10)

// ClaimSetRootKeyDID is a claim of the root key of a merkle tree that goes into the relay.
// this is changed slightly from the version in Iden3 to use our dids instead of there ID type
type ClaimSetRootKeyDID struct {
	// Version is the claim version.
	Version uint32
	// Era is used for labeling epochs.
	Era uint32
	// DID is the DID related to the root key represented in bytes
	DID [32]byte
	// RootKey is the root of the merkle tree.
	RootKey merkletree.Hash
}

// NewClaimSetRootKeyDID returns a ClaimSetRootKey with the given Eth ID and
// merklee tree root key.
func NewClaimSetRootKeyDID(did *didlib.DID, rootKey *merkletree.Hash) (*ClaimSetRootKeyDID, error) {
	if ok := crutils.CheckBigIntArrayInField(merkletree.ElemBytesToBigInts(merkletree.ElemBytes(*rootKey)), crconstants.Q); !ok {
		return nil, errors.New("Elements not in the Finite Field over R")
	}

	didbytes, err := HashDID(did)
	if err != nil {
		return nil, err
	}
	didbytes32 := [32]byte{}
	copy(didbytes32[:], didbytes[:])
	return &ClaimSetRootKeyDID{
		Version: 0,
		Era:     0,
		DID:     didbytes32,
		RootKey: *rootKey,
	}, nil
}

// NewClaimSetRootKeyDIDFromEntry deserializes a ClaimSetRootKeyDID from an Entry.
func NewClaimSetRootKeyDIDFromEntry(e *merkletree.Entry) *ClaimSetRootKeyDID {
	c := &ClaimSetRootKeyDID{}
	_, c.Version = core.GetClaimTypeVersionFromData(&e.Data)

	var era [32 / 8]byte
	copyFromElemBytes(era[:], core.ClaimTypeVersionLen, &e.Data[3])
	c.Era = binary.BigEndian.Uint32(era[:])

	copyFromElemBytes(c.DID[:], 0, &e.Data[2])

	c.RootKey = merkletree.Hash(e.Data[1])
	return c
}

// Entry serializes the claim into an Entry.
func (c ClaimSetRootKeyDID) Entry() *merkletree.Entry {
	e := &merkletree.Entry{}
	core.SetClaimTypeVersion(e, c.Type(), c.Version)
	var era [32 / 8]byte
	binary.BigEndian.PutUint32(era[:], c.Era)
	copyToElemBytes(&e.Data[3], core.ClaimTypeVersionLen, era[:])
	copyToElemBytes(&e.Data[2], 0, c.DID[:])
	e.Data[1] = merkletree.ElemBytes(c.RootKey)
	return e
}

// Type returns the ClaimType of the claim.
func (c *ClaimSetRootKeyDID) Type() core.ClaimType {
	return *ClaimTypeSetRootKeyDID
}
