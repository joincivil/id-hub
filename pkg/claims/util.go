package claims

import (
	"encoding/json"
	"time"

	"github.com/joincivil/id-hub/pkg/claimtypes"
)

// CanonicalizeCredential removes the proof and returns json bytes
func CanonicalizeCredential(cred *claimtypes.ContentCredential) ([]byte, error) {
	temp := &struct {
		Context           []string                            `json:"@context"`
		Type              []claimtypes.CredentialType         `json:"type"`
		CredentialSubject claimtypes.ContentCredentialSubject `json:"credentialSubject"`
		Issuer            string                              `json:"issuer"`
		Holder            string                              `json:"holder,omitempty"`
		CredentialSchema  claimtypes.CredentialSchema         `json:"credentialSchema"`
		IssuanceDate      time.Time                           `json:"issuanceDate"`
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
