package did

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"

	didlib "github.com/ockam-network/did"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
)

const (
	// EthURISchemeMethod is the prefix string for all DIDs in the ethuri DID method
	EthURISchemeMethod = "did:ethuri"
)

// NewService is a convenience function to return a new populated did.Service object
func NewService(persister Persister) *Service {
	return &Service{
		persister: persister,
	}
}

// Service is the service module for DIDs. It is the direct interface for
// managing DIDs and DID documents and should be used when possible.
type Service struct {
	persister Persister
}

// GenerateEthURIDID generates a new EthURI method DID
func (s *Service) GenerateEthURIDID() (*didlib.DID, error) {
	// Generate a new UUID v4
	newUUID := uuid.New()
	didStr := fmt.Sprintf("%s:%s", EthURISchemeMethod, newUUID.String())
	return didlib.Parse(didStr)
}

// InitializeNewDocument generates a simple version of a DID document given
// the DID and an initial public key.
func (s *Service) InitializeNewDocument(did *didlib.DID, firstPK *DocPublicKey) (*Document, error) {
	created := time.Now().UTC()
	updated := time.Now().UTC()

	doc := &Document{
		Context:         DefaultDIDContextV1,
		ID:              *did,
		Controller:      did,
		PublicKeys:      []DocPublicKey{},
		Authentications: []DocAuthenicationWrapper{},
		Created:         &created,
		Updated:         &updated,
	}

	err := doc.AddPublicKey(firstPK.SetIDFragment(doc.NextKeyFragment()), true)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// GetDocument retrieves the DID document given the DID as a string id
// If document is not found, will return a nil Document.
func (s *Service) GetDocument(did string) (*Document, error) {
	d, err := didlib.Parse(did)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing did for get document")
	}

	return s.GetDocumentFromDID(d)
}

// GetDocumentFromDID retrieves the DID document given the DID as a DID object
// If document is not found, will return a nil Document.
func (s *Service) GetDocumentFromDID(did *didlib.DID) (*Document, error) {
	doc, err := s.persister.GetDocument(did)
	if err != nil {
		if err == cpersist.ErrPersisterNoResults {
			return nil, nil
		}
		return nil, errors.Wrap(err, "error getting document from did")
	}
	return doc, nil
}

// SaveDocument saves the DID document given the DID as a string id
func (s *Service) SaveDocument(doc *Document) error {
	return s.persister.SaveDocument(doc)
}
