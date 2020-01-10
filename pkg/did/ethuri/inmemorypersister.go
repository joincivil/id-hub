package ethuri

import (
	"github.com/joincivil/id-hub/pkg/did"
	didlib "github.com/ockam-network/did"
)

// InMemoryPersister is a persister that stores and get did documents in memory
// Mainly used for testing.
type InMemoryPersister struct {
	store map[string]*did.Document
}

// GetDocument retrieves a DID document from the given DID
func (p *InMemoryPersister) GetDocument(d *didlib.DID) (*did.Document, error) {
	if p.store == nil {
		p.store = map[string]*did.Document{}
	}
	theDID := did.MethodIDOnly(d)
	doc, ok := p.store[theDID]
	if !ok {
		return nil, nil
	}
	return doc, nil
}

// SaveDocument saves a DID document
func (p *InMemoryPersister) SaveDocument(doc *did.Document) error {
	if p.store == nil {
		p.store = map[string]*did.Document{}
	}
	theDID := did.MethodIDOnly(&doc.ID)
	p.store[theDID] = doc
	return nil
}
