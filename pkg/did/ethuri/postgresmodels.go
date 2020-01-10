package ethuri

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/pkg/errors"
)

// PostgresDocument is the GORM model for storing the serialized DID document to DID
// mapping
type PostgresDocument struct {
	DID       string         `gorm:"column:did;primary_key"`
	Document  postgres.Jsonb `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// TableName sets the tablename for PostgresDocuments
func (PostgresDocument) TableName() string {
	return "dids"
}

// ToDocument returns the did.Document from this PostgresDocument
func (p *PostgresDocument) ToDocument() (*did.Document, error) {
	if p.Document.RawMessage == nil || len(p.Document.RawMessage) == 0 {
		return nil, errors.New("no jsonb document data to unmarshal")
	}

	doc := &did.Document{}
	err := json.Unmarshal(p.Document.RawMessage, doc)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling db postgres doc")
	}

	return doc, nil
}

// FromDocument sets up this PostgresDocument with data from the given did.Document
func (p *PostgresDocument) FromDocument(doc *did.Document) error {
	bys, err := json.Marshal(doc)
	if err != nil {
		return errors.Wrap(err, "error marshalling doc for postgres doc")
	}

	p.DID = doc.ID.String()
	p.Document.RawMessage = bys
	// NOTE(PN): Let GORM set the CreatedAt, UpdatedAt, DeletedAt for this object
	// which is different than the document's internal created, updated values

	return nil
}
