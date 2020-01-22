package did

import (
	"sync"

	log "github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/Jeffail/tunny"
	didlib "github.com/ockam-network/did"
)

const (
	numWorkers = 3
)

// NewService is a convenience function to return a new populated did.Service object
func NewService(resolvers []Resolver) *Service {
	service := &Service{
		resolvers: resolvers,
	}
	pool := tunny.NewFunc(
		numWorkers,
		service.resolveProcess,
	)
	return &Service{
		resolvers:   resolvers,
		resProcPool: pool,
	}
}

// Service is the service module for DIDs. It is the direct interface for
// managing DIDs and DID documents and should be used when possible.
type Service struct {
	resolvers   []Resolver
	resProcPool *tunny.Pool
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

type resolveParams struct {
	r Resolver
	d *didlib.DID
}

func (s *Service) resolveProcess(payload interface{}) interface{} {
	p, ok := payload.(resolveParams)
	if !ok {
		log.Errorf("Payload is not resolveParams: %v", payload)
		return nil
	}

	doc, err := p.r.Resolve(p.d)
	if err == nil && doc != nil {
		log.Infof("resolved: %T, did: %v", p.r, p.d.String())
		return doc
	}

	log.Infof("unresolved: %T, did: %v err: %v", p.r, p.d.String(), err)
	return nil
}

// resolve runs through the given slice of Resolvers, spawns them in different
// goroutines and returns the first valid result.
func (s *Service) resolve(d *didlib.DID) (*Document, error) {
	results := make([]*Document, len(s.resolvers))
	var wg sync.WaitGroup

	for ind, r := range s.resolvers {
		wg.Add(1)
		go func(r Resolver, d *didlib.DID, ind int) {
			defer wg.Done()
			result := s.resProcPool.Process(resolveParams{
				r: r,
				d: d,
			})

			doc, ok := result.(*Document)
			if ok {
				results[ind] = doc
			} else {
				results[ind] = nil
			}
		}(r, d, ind)
	}

	wg.Wait()

	// Return the first non-nil result
	for _, doc := range results {
		if doc != nil {
			return doc, nil
		}
	}

	return nil, ErrResolverDIDNotFound
}
