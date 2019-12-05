package claimtypes

import (
	"encoding/json"
	"time"

	"github.com/joincivil/id-hub/pkg/linkeddata"
)

// LicenseCredential a credential that represents the righ to use
// some content granted by the owner
// https://www.w3.org/TR/vc-data-model/#basic-concepts
type LicenseCredential struct {
	Context           []string         `json:"@context"`
	Type              []CredentialType `json:"type"`
	CredentialSubject interface{}      `json:"credentialSubject"`
	Issuer            string           `json:"issuer"`
	Holder            string           `json:"holder,omitempty"`
	CredentialSchema  CredentialSchema `json:"credentialSchema"`
	IssuanceDate      time.Time        `json:"issuanceDate"`
	ExpirationDate    time.Time        `json:"expirationDate"`
	Proof             interface{}      `json:"proof,omitempty"`
}

// ContentSubject represents the content being licensed
// https://www.w3.org/TR/vc-data-model/#credential-subject
type ContentSubject struct {
	ID    string `json:"id"`
	Owner string `json:"owner"`
}

// LicenserSubject is the ententity licensing the content
// https://www.w3.org/TR/vc-data-model/#credential-subject
type LicenserSubject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FindLinkedDataProof returns the first the linked data proof in the proof slice
func (c *LicenseCredential) FindLinkedDataProof() (*linkeddata.Proof, error) {
	return FindLinkedDataProof(c.Proof)
}

// CanonicalizeCredential converts the cred into a predictable json format for signing
func (c *LicenseCredential) CanonicalizeCredential() ([]byte, error) {
	temp := &struct {
		Context           []string         `json:"@context"`
		Type              []CredentialType `json:"type"`
		CredentialSubject interface{}      `json:"credentialSubject"`
		Issuer            string           `json:"issuer"`
		Holder            string           `json:"holder,omitempty"`
		CredentialSchema  CredentialSchema `json:"credentialSchema"`
		IssuanceDate      time.Time        `json:"issuanceDate"`
		ExpirationDate    time.Time        `json:"expirationDate"`
	}{
		Context:           c.Context,
		Type:              c.Type,
		CredentialSubject: c.CredentialSubject,
		Issuer:            c.Issuer,
		Holder:            c.Holder,
		CredentialSchema:  c.CredentialSchema,
		IssuanceDate:      c.IssuanceDate,
		ExpirationDate:    c.ExpirationDate,
	}
	return json.Marshal(temp)
}
