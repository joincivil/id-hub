package claims

import (
	"encoding/json"
	"time"

	"github.com/joincivil/id-hub/pkg/claimsstore"
)

// CanonicalizeCredential removes the proof and returns json bytes
func CanonicalizeCredential(cred *claimsstore.ContentCredential) ([]byte, error) {
	temp := &struct {
		Context           []string                             `json:"@context"`
		Type              []claimsstore.CredentialType         `json:"type"`
		CredentialSubject claimsstore.ContentCredentialSubject `json:"credentialSubject"`
		Issuer            string                               `json:"issuer"`
		Holder            string                               `json:"holder,omitempty"`
		CredentialSchema  claimsstore.CredentialSchema         `json:"credentialSchema"`
		IssuanceDate      time.Time                            `json:"issuanceDate"`
	}{
		Context:           cred.Context,
		Type:              cred.Type,
		CredentialSubject: cred.CredentialSubject,
		Issuer:            cred.Issuer,
		Holder:            cred.Holder,
		CredentialSchema:  cred.CredentialSchema,
		IssuanceDate:      cred.IssuanceDate,
	}
	return json.Marshal(temp)
}
