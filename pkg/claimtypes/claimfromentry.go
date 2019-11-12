package claimtypes

import (
	"github.com/iden3/go-iden3-core/core"
	"github.com/iden3/go-iden3-core/merkletree"
)

// NewClaimFromEntry extends iden3 NewClaimFromEntry with our claim types
func NewClaimFromEntry(entry *merkletree.Entry) (merkletree.Claim, error) {
	claim, err := core.NewClaimFromEntry(entry)
	if err == core.ErrInvalidClaimType {
		claimType, _ := core.GetClaimTypeVersion(entry)
		if claimType == *ClaimTypeSetRootKeyDID {
			claim = NewClaimSetRootKeyDIDFromEntry(entry)
			return claim, nil
		}
		if claimType == *ClaimTypeRegisteredDocument {
			claim := NewClaimRegisteredDocumentFromEntry(entry)
			return claim, nil
		}
		return nil, core.ErrInvalidClaimType
	}

	return claim, err
}
