package graphql

import (
	"encoding/json"

	log "github.com/golang/glog"
	"github.com/joincivil/id-hub/pkg/did"
)

// DidGetResponse represents the GraphQL response for DidGet
type DidGetResponse struct {
	Doc *did.Document `json:"doc"`
}

// DocRaw returns the raw JSON string for the docRaw field
func (d *DidGetResponse) DocRaw() *string {
	bys, err := json.Marshal(d.Doc)
	if err != nil {
		log.Errorf("Error marshalling doc: err: %v", err)
		return nil
	}
	docRaw := string(bys)
	return &docRaw
}

// DidSaveResponse represents the GraphQL response for DidSave
type DidSaveResponse struct {
	Doc *did.Document `json:"doc"`
}

// DocRaw returns the raw JSON string for the docRaw field
func (d *DidSaveResponse) DocRaw() *string {
	bys, err := json.Marshal(d.Doc)
	if err != nil {
		log.Errorf("Error marshalling doc: err: %v", err)
		return nil
	}
	docRaw := string(bys)
	return &docRaw
}
