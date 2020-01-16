package ethuri

import (
	"strings"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
	didlib "github.com/ockam-network/did"
	"github.com/pkg/errors"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

// Implements the did.Resolver interface, so can be passed into did.Service

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

// Resolve implements the did.Resolver interface and returns the did document of a
// given DID for the ethuri method.
func (s *Service) Resolve(d *didlib.DID) (*did.Document, error) {
	if !s.IsEthURI(d.String()) {
		return nil, did.ErrResolverDIDNotFound
	}

	doc, err := s.GetDocumentFromDID(d)
	if doc == nil && err == nil {
		return nil, did.ErrResolverDIDNotFound
	}

	return doc, err
}

// IsEthURI is a quick check to see if a did is for ethuri
func (s *Service) IsEthURI(d string) bool {
	return strings.HasPrefix(d, EthURISchemeMethod)
}

// CreateOrUpdateParams are input params for CreateOrUpdateDocument
type CreateOrUpdateParams struct {
	Did              *string
	PublicKeys       []did.DocPublicKey
	Auths            []did.DocAuthenicationWrapper
	Services         []did.DocService
	Proof            *linkeddata.Proof
	KeepKeyFragments bool
}

// GetDocument retrieves the DID document given the DID as a string id
// If document is not found, will return a nil Document.
func (s *Service) GetDocument(did string) (*did.Document, error) {
	d, err := didlib.Parse(did)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing did for get document")
	}

	return s.GetDocumentFromDID(d)
}

// GetDocumentFromDID retrieves the DID document given the DID as a DID object
// If document is not found, will return a nil Document.
func (s *Service) GetDocumentFromDID(did *didlib.DID) (*did.Document, error) {
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
func (s *Service) SaveDocument(doc *did.Document) error {
	return s.persister.SaveDocument(doc)
}

// CreateOrUpdateDocument will create a new document or update an existing one given
// the params in CreateOrUpdateParams.  If did is given and valid, will attempt to
// retrieve the existing did and document and add any new data to the document
// If no did is given, it will create a new document with a new DID and the given data.
// In both cases it will persist to store.
func (s *Service) CreateOrUpdateDocument(p *CreateOrUpdateParams) (*did.Document, error) {
	var doc *did.Document
	var err error

	if p.Did != nil {
		doc, err = s.updateDocumentFromParams(p)
		if err != nil {
			return nil, errors.Wrap(err, "error updating new document")
		}

	} else {
		doc, err = s.createNewDocumentFromParams(p)
		if err != nil {
			return nil, errors.Wrap(err, "error creating new document")
		}
	}

	err = s.SaveDocument(doc)
	if err != nil {
		return nil, errors.Wrap(err, "error storing new did")
	}

	return doc, nil
}

func (s *Service) updateDocumentFromParams(p *CreateOrUpdateParams) (*did.Document, error) {
	var doc *did.Document
	var err error

	// Validate the DID
	if !did.ValidDid(*p.Did) {
		return nil, errors.New("did is invalid")
	}

	// Try to retrieve the DID document
	doc, err = s.GetDocument(*p.Did)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get document for did")
	}

	if doc == nil {
		return nil, errors.New("no did found to update")
	}

	for _, pk := range p.PublicKeys {
		err = doc.AddPublicKey(&pk, false, !p.KeepKeyFragments)
		if err != nil {
			return nil, errors.Wrap(err, "unable to add public key")
		}
	}

	for _, auth := range p.Auths {
		err = doc.AddAuthentication(&auth, !p.KeepKeyFragments)
		if err != nil {
			return nil, errors.Wrap(err, "unable to add authentication")
		}
	}

	for _, srv := range p.Services {
		err = doc.AddService(&srv)
		if err != nil {
			return nil, errors.Wrap(err, "unable to add service")
		}
	}

	// Update proof
	doc.Proof = p.Proof

	return doc, nil
}

func (s *Service) createNewDocumentFromParams(p *CreateOrUpdateParams) (*did.Document, error) {
	if len(p.PublicKeys) == 0 {
		return nil, errors.New("at least one public key required")
	}

	// Generate a new document with the first key
	doc, err := GenerateNewDocument(&p.PublicKeys[0], true, !p.KeepKeyFragments)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate new did document")
	}

	// Add rest of keys if more than one
	for _, pk := range p.PublicKeys[1:] {
		err = doc.AddPublicKey(&pk, false, !p.KeepKeyFragments)
		if err != nil {
			return nil, errors.Wrap(err, "unable to add public key")
		}
	}

	// Add the auths
	for _, auth := range p.Auths {
		err = doc.AddAuthentication(&auth, !p.KeepKeyFragments)
		if err != nil {
			return nil, errors.Wrap(err, "unable to add auth")
		}
	}

	// Add the services
	for _, srv := range p.Services {
		err = doc.AddService(&srv)
		if err != nil {
			return nil, errors.Wrap(err, "unable to add service")
		}
	}

	doc.Proof = p.Proof

	return doc, nil
}
