package did

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	didlib "github.com/ockam-network/did"
)

// https://github.com/ockam-network/did as base DID parser/handler.

// Document is the base definition of a DID document
// https://w3c-ccg.github.io/did-spec/#did-documents
type Document struct {
	Context        string                    `json:"@context"`
	ID             didlib.DID                `json:"id"`
	Controller     *didlib.DID               `json:"controller,omitempty"`
	PublicKeys     []DocPublicKey            `json:"publicKey"`
	Authentication []DocAuthenicationWrapper `json:"authentication,omitempty"`
	Service        []DocService              `json:"service,omitempty"`
	Created        time.Time                 `json:"created,omitempty"`
	Updated        time.Time                 `json:"updated,omitempty"`
	Proof          LinkedDataSignature       `json:"proof,omitempty"`
}

func (d Document) String() string {
	buf := bytes.NewBufferString("Document: ")
	buf.WriteString(fmt.Sprintf("id: %v, ", d.ID.String()))
	if d.Controller != nil {
		buf.WriteString(fmt.Sprintf("controller: %v, ", d.Controller.String()))
	}

	buf.WriteString(fmt.Sprintf("num keys: %v, ", len(d.PublicKeys)))
	buf.WriteString(fmt.Sprintf("num auth: %v, ", len(d.Authentication)))

	if d.Proof != (LinkedDataSignature{}) {
		buf.WriteString(fmt.Sprintf("proof: %v, ", d.Proof.SignatureValue))
	}
	if d.Created != (time.Time{}) {
		buf.WriteString(fmt.Sprintf("created: %v, ", d.Created))
	}
	if d.Updated != (time.Time{}) {
		buf.WriteString(fmt.Sprintf("updated: %v, ", d.Updated))
	}

	return buf.String()
}

// UnmarshalJSON implements the Unmarshaller interface for Document
func (d *Document) UnmarshalJSON(b []byte) error {
	type docAlias Document
	aux := &struct {
		ID         string `json:"id"`
		Controller string `json:"controller,omitempty"`
		*docAlias
	}{
		docAlias: (*docAlias)(d),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal document")
	}

	// Set the DID as a struct
	id, err := didlib.Parse(aux.ID)
	if err != nil {
		return errors.Wrap(err, "unable to parse did for document")
	}
	d.ID = *id

	// Set the controller as a struct
	if aux.Controller != "" {
		controller, err := didlib.Parse(aux.Controller)
		if err != nil {
			return errors.Wrap(err, "unable to parse controller for document")
		}
		d.Controller = controller
	}

	return nil
}

// MarshalJSON implements the Marshaller interface for Document
func (d *Document) MarshalJSON() ([]byte, error) {
	type docAlias Document
	aux := &struct {
		ID         string `json:"id"`
		Controller string `json:"controller,omitempty"`
		*docAlias
	}{
		ID:       d.ID.String(),
		docAlias: (*docAlias)(d),
	}

	if d.Controller != nil {
		aux.Controller = d.Controller.String()
	}

	return json.Marshal(aux)
}

// DocPublicKey defines a publickey within a DID document
type DocPublicKey struct {
	ID                 didlib.DID `json:"id"`
	Type               string     `json:"type"`
	Controller         string     `json:"controller"`
	PublicKeyPem       string     `json:"publicKeyPem,omitempty"`
	PublicKeyJwk       string     `json:"publicKeyJwk,omitempty"`
	PublicKeyHex       string     `json:"publicKeyHex,omitempty"`
	PublicKeyBase64    string     `json:"publicKeyBase64,omitempty"`
	PublicKeyBase58    string     `json:"publicKeyBase58,omitempty"`
	PublicKeyMultibase string     `json:"publicKeyMultibase,omitempty"`
	EthereumAddress    string     `json:"ethereumAddress,omitempty"`
}

// UnmarshalJSON implements the Unmarshaler interface for DocPublicKey
func (p *DocPublicKey) UnmarshalJSON(b []byte) error {
	type pkAlias DocPublicKey
	aux := &struct {
		ID string `json:"id"`
		*pkAlias
	}{
		pkAlias: (*pkAlias)(p),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal public key")
	}

	// Set the DID as a struct
	id, err := didlib.Parse(aux.ID)
	if err != nil {
		return errors.Wrap(err, "unable to parse did for public key")
	}
	p.ID = *id

	return nil
}

// MarshalJSON implements the Marshaler interface for DocPublicKey
func (p *DocPublicKey) MarshalJSON() ([]byte, error) {
	type pkAlias DocPublicKey
	aux := &struct {
		ID string `json:"id"`
		*pkAlias
	}{
		ID:      p.ID.String(),
		pkAlias: (*pkAlias)(p),
	}

	return json.Marshal(aux)
}

// DocAuthenicationWrapper allows us to handle two different types for an authentication
// value.  This can either be an ID to a public key or a public key.
type DocAuthenicationWrapper struct {
	DocPublicKey
	Partial bool `json:"-"`
}

// UnmarshalJSON implements the Unmarshaler interface for
func (a *DocAuthenicationWrapper) UnmarshalJSON(b []byte) error {
	type awAlias DocAuthenicationWrapper
	aux := &struct {
		ID string `json:"id"`
		*awAlias
	}{
		awAlias: (*awAlias)(a),
	}

	// If it is a JSON string for a public key object
	err := json.Unmarshal(b, &aux)

	// If no err, then it should have unmarshaled properly
	// DocPublicKey.UnmarshalJSON will also run and that will convert the ID properly.
	if err == nil {
		return nil
	}

	// If it is a DID string
	// Strip out any whitespace or quotes
	id := strings.Trim(string(b), "\" ")
	d, err := didlib.Parse(id)
	if err != nil {
		return errors.Wrapf(err, "unable to parse auth did: %v", string(b))
	}

	a.ID = *d
	a.Partial = true

	return nil
}

// MarshalJSON implements the Marshaler interface for
func (a *DocAuthenicationWrapper) MarshalJSON() ([]byte, error) {
	if a.Partial {
		// Need to wrap in quotes to make it valid as a JSON string
		return []byte(fmt.Sprintf("\"%v\"", a.ID.String())), nil
	}

	type awAlias DocAuthenicationWrapper
	aux := &struct {
		ID string `json:"id"`
		*awAlias
	}{
		ID:      a.ID.String(),
		awAlias: (*awAlias)(a),
	}

	return json.Marshal(aux)
}

// DocService defines a service endpoint within a DID document
type DocService struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	// DocServiceEndpoint could be a JSON-LD object or a URI
	// https://github.com/piprate/json-gold
	// string or map[string]interface{}
	ServiceEndpoint interface{} `json:"serviceEndpoint"`

	// DocServiceEndpoint values stored here as the correct type
	// Use these to access the values for DocServiceEndpoint
	ServiceEndpointURI *string                `json:"-"`
	ServiceEndpointLD  map[string]interface{} `json:"-"`
}

// UnmarshalJSON implements the Unmarshaler interface for DocService
func (s *DocService) UnmarshalJSON(b []byte) error {
	type alias DocService
	aux := &struct {
		*alias
	}{
		alias: (*alias)(s),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal public key")
	}

	// // Validate types for service endpoint.  Can either be a string (URI)
	// // or a JSON-LD object
	switch val := s.ServiceEndpoint.(type) {
	case string:
		// valid type for URIs
		s.ServiceEndpointURI = &val
	case map[string]interface{}:
		// valid type
		// TODO: do more for validation of JSON-LD
		s.ServiceEndpointLD = val
	default:
		return errors.New("invalid type for serviceendpoint value")
	}

	return nil
}

// LinkedDataSignature defines a linked data signature object
// Spec: https://w3c-dvcg.github.io/ld-signatures/#linked-data-signature-overview
type LinkedDataSignature struct {
	Type           string    `json:"type"`
	Creator        string    `json:"creator"`
	Created        time.Time `json:"created"`
	SignatureValue string    `json:"signatureValue"`
	Domain         string    `json:"domain,omitempty"`
	Nonce          string    `json:"nonce,omitempty"`
}
