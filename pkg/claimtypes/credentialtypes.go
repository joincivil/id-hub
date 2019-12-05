package claimtypes

import "github.com/joincivil/id-hub/pkg/linkeddata"

// CredentialType is a non-exclusive type for a credential
type CredentialType string

const (
	// VerifiableCredentialType standard credential type
	VerifiableCredentialType CredentialType = "VerifiableCredential"
	// ContentCredentialType marks the credential as a content credential
	ContentCredentialType CredentialType = "ContentCredential"
	// LicenseCredentialType marks the credential as a license credential
	LicenseCredentialType CredentialType = "LicenseCredential"
)

// Credential is a common interface for credentials
type Credential interface {
	CanonicalizeCredential() ([]byte, error)
	FindLinkedDataProof() (*linkeddata.Proof, error)
}
