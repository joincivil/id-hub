package did

import (
	log "github.com/golang/glog"
	"github.com/pkg/errors"

	didlib "github.com/ockam-network/did"
	// "github.com/joincivil/id-hub/pkg/linkeddata"
)

// NewService is a convenience function to return a new populated did.Service object
func NewService(resolvers []Resolver) *Service {
	return &Service{
		resolvers: resolvers,
	}
}

// Service is the service module for DIDs. It is the direct interface for
// managing DIDs and DID documents and should be used when possible.
type Service struct {
	resolvers []Resolver
}

// GetDocument retrieves the DID document given the DID as a string id
// If document is not found, will return a nil Document.
func (s *Service) GetDocument(did string) (*Document, error) {
	d, err := didlib.Parse(did)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing did for get document")
	}

	return s.resolve(d)
}

// GetDocumentFromDID retrieves the DID document given the DID as a DID object
// If document is not found, will return a nil Document.
func (s *Service) GetDocumentFromDID(did *didlib.DID) (*Document, error) {
	return s.resolve(did)
}

// GetKeyFromDIDDocument returns a public key document from a did with a fragment if it can be found
// errors if fragment is empty
func (s *Service) GetKeyFromDIDDocument(did *didlib.DID) (*DocPublicKey, error) {
	// Make copy to avoid side-effects to altering by reference
	d := CopyDID(did)
	fragment := d.Fragment
	if fragment == "" {
		return nil, errors.New("no fragment on did")
	}
	d.Fragment = ""

	doc, err := s.GetDocumentFromDID(d)
	if err != nil {
		return nil, err
	}

	if doc == nil {
		return nil, errors.New("no did document found")
	}

	return doc.GetPublicKeyFromFragment(fragment)
}

// resolve runs through the given slice of Resolvers and returns the
// first valid result. To later improve performance, can spawn goroutines in a pool
// and query the resolvers simultaneously, returning the "best" answer.
func (s *Service) resolve(d *didlib.DID) (*Document, error) {
	var doc *Document
	var err error

	for _, r := range s.resolvers {
		doc, err = r.Resolve(d)
		if err == nil && doc != nil {
			return doc, nil
		}
		log.Infof("%v -> err: %v", d.String(), err)
	}

	return nil, err
}
