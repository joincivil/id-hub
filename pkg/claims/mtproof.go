package claims

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-core/merkletree"
)

// MTProof is
type MTProof struct {
	ExistsInDIDMTProof     string // HEX
	NotRevokedInDIDMTProof string // HEX
	DIDRootExistsProof     string // HEX
	DIDRootExistsVersion   uint32 // The version of the claim in the tree, this is needed to verify the proof
	BlockNumber            int64
	ContractAddress        common.Address
	TXHash                 common.Hash
	Root                   merkletree.Hash
	DIDRoot                merkletree.Hash
	CommitterAddress       common.Address
	DID                    string
}
