package ethuri

import (
	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
	didlib "github.com/ockam-network/did"

	"github.com/joincivil/id-hub/pkg/did"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
)

// NewPostgresPersister is a convenience function to return a populated PostgresPersister
func NewPostgresPersister(db *gorm.DB) *PostgresPersister {
	return &PostgresPersister{
		db: db,
	}
}

// PostgresPersister is the Postgresql implementation of the DID persister
type PostgresPersister struct {
	db *gorm.DB
}

// GetDocument retrieves a DID document from the given DID
func (p *PostgresPersister) GetDocument(d *didlib.DID) (*did.Document, error) {
	if d == nil {
		return nil, errors.New("nil did for get document")
	}

	theDID := did.MethodIDOnly(d)

	doc := &PostgresDocument{}
	err := p.db.Where(&PostgresDocument{DID: theDID}).First(doc).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, cpersist.ErrPersisterNoResults
		}
		return nil, errors.Wrap(err, "error getting did document")
	}

	return doc.ToDocument()
}

// SaveDocument saves a DID document with the given DID
func (p *PostgresPersister) SaveDocument(doc *did.Document) error {
	dbdoc := &PostgresDocument{}
	err := dbdoc.FromDocument(doc)
	if err != nil {
		return errors.Wrap(err, "error setting up db doc with document")
	}

	updated := &PostgresDocument{}
	err = p.db.Where(&PostgresDocument{DID: did.MethodIDOnly(&doc.ID)}).
		Assign(&PostgresDocument{Document: dbdoc.Document}).
		FirstOrCreate(updated).Error
	if err != nil {
		return errors.Wrap(err, "error saving up db doc")
	}

	return nil
}
