package did

import (
	didlib "github.com/ockam-network/did"

	"github.com/pkg/errors"
)

// Resolver interface that defines a DID resolver
type Resolver interface {
	// Resolve returns the DID document given the DID
	Resolve(d *didlib.DID) (*Document, error)
}

var (
	// ErrResolverDIDNotFound error indicating that DID not found by resolver
	ErrResolverDIDNotFound = errors.New("did doc not found")

	// ErrResolverCacheDIDNotFound error indicating that DID not found in cache
	ErrResolverCacheDIDNotFound = errors.New("did doc not found in cache")
)

// ResolverCache interface defines a DID document cache for the resolver
type ResolverCache interface {
	Get(d *didlib.DID) (*Document, error)
	Set(d *didlib.DID, doc *Document) error
}
