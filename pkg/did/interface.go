package did

import (
	didlib "github.com/ockam-network/did"
)

// Persister is the interface of storing and retrieving DID documents
// for the ID hub. Implement this interface with different backing stores.
type Persister interface {
	// GetDocument retrieves a DID document from the given DID
	GetDocument(d *didlib.DID) (*Document, error)
	// SaveDocument saves a DID document
	SaveDocument(doc *Document) error
}
