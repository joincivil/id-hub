package claims

import (
	"encoding/json"
	"time"

	"github.com/iden3/go-iden3-core/merkletree"
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

// these are methods copied from iden3 that they made private but are very useful

// copyToElemBytes copies the src slice forwards to e, ending at -start of
// e.  This function will panic if src doesn't fit into len(e)-start.
func copyToElemBytes(e *merkletree.ElemBytes, start int, src []byte) {
	copy(e[merkletree.ElemBytesLen-start-len(src):], src)
}

// copyFromElemBytes copies from e to dst, ending at -start of e and going
// forwards.  This function will panic if len(e)-start is smaller than
// len(dst).
func copyFromElemBytes(dst []byte, start int, e *merkletree.ElemBytes) {
	copy(dst, e[merkletree.ElemBytesLen-start-len(dst):])
}
