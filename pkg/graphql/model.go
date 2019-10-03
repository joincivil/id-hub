package graphql

import (
	"encoding/json"

	log "github.com/golang/glog"
	"github.com/joincivil/id-hub/pkg/claimsstore"
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

// ClaimGetResponse represents the GraphQL response for ClaimGet
type ClaimGetResponse struct {
	Claims []*claimsstore.ContentCredential `json:"claims"`
}

// ClaimsRaw returns the raw JSON string for the list of claims
func (d *ClaimGetResponse) ClaimsRaw() []string {
	clms := make([]string, len(d.Claims))

	for ind, c := range d.Claims {
		bys, err := json.Marshal(c)
		if err != nil {
			log.Errorf("Error marshalling claim: err: %v", err)
			continue
		}
		clms[ind] = string(bys)
	}

	return clms
}

// ClaimSaveResponse represents the GraphQL response for ClaimSave
// type ClaimSaveResponse struct {
// 	Claim *claimsstore.ContentCredential `json:"claim"`
// }

// ClaimRaw returns the raw JSON string of the claim
// func (d *ClaimSaveResponse) ClaimRaw() *string {
// 	bys, err := json.Marshal(d.Claims)
// 	if err != nil {
// 		log.Errorf("Error marshalling claim: err: %v", err)
// 		return nil
// 	}
// 	claimRaw := string(bys)
// 	return &claimRaw
// }
