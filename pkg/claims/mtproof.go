package claims

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-core/merkletree"
)

// MTProof is
type MTProof struct {
	ExistsInDIDMTProof     string          `json:"entryExistsInIssuerTree"`     // HEX
	NotRevokedInDIDMTProof string          `json:"entryNotRevokedInIssuerTree"` // HEX
	DIDRootExistsProof     string          `json:"issuerRootExistsInRelayTree"` // HEX
	DIDRootExistsVersion   uint32          `json:"issuerRootVersion"`           // The version of the claim in the tree, this is needed to verify the proof
	BlockNumber            int64           `json:"blockNumber"`
	ContractAddress        common.Address  `json:"contractAddress"`
	TXHash                 common.Hash     `json:"txHash"`
	Root                   merkletree.Hash `json:"relayTreeRoot"`
	DIDRoot                merkletree.Hash `json:"issuerTreeRoot"`
	CommitterAddress       common.Address  `json:"relayAddress"`
	DID                    string          `json:"issuer"`
}
