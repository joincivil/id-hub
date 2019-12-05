package claimtypes

import (
	"encoding/json"
	"time"

	"github.com/joincivil/go-common/pkg/article"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

// ContentCredential a credential that claims a piece of content
// https://www.w3.org/TR/vc-data-model/#basic-concepts
type ContentCredential struct {
	Context           []string                 `json:"@context"`
	Type              []CredentialType         `json:"type"`
	CredentialSubject ContentCredentialSubject `json:"credentialSubject"`
	Issuer            string                   `json:"issuer"`
	Holder            string                   `json:"holder,omitempty"`
	CredentialSchema  CredentialSchema         `json:"credentialSchema"`
	IssuanceDate      time.Time                `json:"issuanceDate"`
	Proof             interface{}              `json:"proof,omitempty"`
}

// ContentCredentialSubject the datatype for claiming a piece of content
// https://www.w3.org/TR/vc-data-model/#credential-subject
type ContentCredentialSubject struct {
	ID       string           `json:"id"`
	Metadata article.Metadata `json:"metadata"`
}

// CredentialSchema for specifying schemas for a credential type
// https://www.w3.org/TR/vc-data-model/#data-schemas
type CredentialSchema struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// FindLinkedDataProof returns the first the linked data proof in the proof slice
func (c *ContentCredential) FindLinkedDataProof() (*linkeddata.Proof, error) {
	return FindLinkedDataProof(c.Proof)
}

// CanonicalizeCredential removes the proof and returns json bytes
func (c *ContentCredential) CanonicalizeCredential() ([]byte, error) {
	temp := &struct {
		Context           []string                 `json:"@context"`
		Type              []CredentialType         `json:"type"`
		CredentialSubject ContentCredentialSubject `json:"credentialSubject"`
		Issuer            string                   `json:"issuer"`
		Holder            string                   `json:"holder,omitempty"`
		CredentialSchema  CredentialSchema         `json:"credentialSchema"`
		IssuanceDate      time.Time                `json:"issuanceDate"`
	}{
		Context:           c.Context,
		Type:              c.Type,
		CredentialSubject: c.CredentialSubject,
		Issuer:            c.Issuer,
		Holder:            c.Holder,
		CredentialSchema:  c.CredentialSchema,
		IssuanceDate:      c.IssuanceDate,
	}
	return json.Marshal(temp)
}
