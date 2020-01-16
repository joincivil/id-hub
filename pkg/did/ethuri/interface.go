package ethuri

import (
	"github.com/joincivil/id-hub/pkg/did"
	didlib "github.com/ockam-network/did"
)

// Persister is the interface of storing and retrieving DID documents
// for the ID hub. Implement this interface with different backing stores.
type Persister interface {
	// GetDocument retrieves a DID document from the given DID
	GetDocument(d *didlib.DID) (*did.Document, error)
	// SaveDocument saves a DID document
	SaveDocument(doc *did.Document) error
}
