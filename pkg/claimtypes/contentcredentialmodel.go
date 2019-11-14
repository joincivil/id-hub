package claimtypes

import (
	"encoding/json"
	"errors"
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
	Proof             []interface{}            `json:"proof,omitempty"`
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
func FindLinkedDataProof(proofs []interface{}) (*linkeddata.Proof, error) {
	for _, v := range proofs {
		switch tv := v.(type) {
		case *linkeddata.Proof:
			return tv, nil
		case linkeddata.Proof:
			return &tv, nil
		case map[string]interface{}:
			t, ok := tv["type"]
			if ok && t == string(linkeddata.SuiteTypeSecp256k1Signature) {
				ld := &linkeddata.Proof{}
				js, err := json.Marshal(tv)

				if err != nil {
					return nil, err
				}

				err = json.Unmarshal(js, ld)
				if err != nil {
					return nil, err
				}
				return ld, nil
			}
			return nil, errors.New("proofs array didn't contain a linked data proof")
		}
	}
	return nil, errors.New("proofs array didn't contain a linked data proof")
}