package did

import (
	"github.com/ockam-network/did"
)

// Resolver interface that defines a DID resolver
type Resolver interface {
	// Resolve returns the DID document given the DID
	Resolve(d *did.DID) (*Document, error)
}
