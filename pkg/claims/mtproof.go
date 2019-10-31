package claims

import (
	"github.com/ethereum/go-ethereum/common"
)

// MTProof is
type MTProof struct {
	ExistsInDIDMTProof     string // HEX
	NotRevokedInDIDMTProof string // HEX
	DIDRootExistsProof     string
	BlockNumber            int64
	BlockTimestamp         int64
	ContractAddress        common.Address
	TXHash                 common.Hash
}
