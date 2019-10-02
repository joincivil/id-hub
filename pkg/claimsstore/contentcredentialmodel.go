package claimsstore

import (
	"time"

	"github.com/joincivil/go-common/pkg/article"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

// CredentialType is a non-exclusive type for a credential
type CredentialType string

const (
	// VerifiableCredentialType standard credential type
	VerifiableCredentialType CredentialType = "VerifiableCredential"
	// ContentCredentialType marks the credential as a content credential
	ContentCredentialType CredentialType = "ContentCredential"
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
	Proof             linkeddata.Proof         `json:"proof,omitempty"`
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
