package did

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/golang/glog"

	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/id-hub/pkg/utils"
	"github.com/pkg/errors"

	didlib "github.com/ockam-network/did"
)

const (
	// DefaultDIDContextV1 is the default context for DID documents
	DefaultDIDContextV1 = "https://www.w3.org/2019/did/v1"
)

// https://github.com/ockam-network/did as base DID parser/handler.

// Document is the base definition of a DID document
// https://w3c-ccg.github.io/did-spec/#did-documents
type Document struct {
	Context         string                    `json:"@context"`
	ID              didlib.DID                `json:"id"`
	Controller      *didlib.DID               `json:"controller,omitempty"`
	PublicKeys      []DocPublicKey            `json:"publicKey"`
	Authentications []DocAuthenicationWrapper `json:"authentication,omitempty"`
	Services        []DocService              `json:"service,omitempty"`
	Created         *time.Time                `json:"created,omitempty"`
	Updated         *time.Time                `json:"updated,omitempty"`
	Proof           *LinkedDataProof          `json:"proof,omitempty"`
}

func (d Document) String() string {
	buf := bytes.NewBufferString("Document: ")
	buf.WriteString(fmt.Sprintf("id: %v, ", d.ID.String()))
	if d.Controller != nil {
		buf.WriteString(fmt.Sprintf("controller: %v, ", d.Controller.String()))
	}

	buf.WriteString(fmt.Sprintf("num keys: %v, ", len(d.PublicKeys)))
	buf.WriteString(fmt.Sprintf("num auths: %v, ", len(d.Authentications)))
	buf.WriteString(fmt.Sprintf("num services: %v, ", len(d.Services)))

	if d.Proof != nil {
		buf.WriteString(fmt.Sprintf("proof: %v, ", d.Proof.ProofValue))
	}
	if d.Created != nil {
		buf.WriteString(fmt.Sprintf("created: %v, ", d.Created))
	}
	if d.Updated != nil {
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

	if d.Controller != nil && d.Controller.String() != "" {
		aux.Controller = d.Controller.String()
	}

	return json.Marshal(aux)
}

// AddPublicKey adds another public key.
// If addRefToAuth is true, also adds a reference to the key in the authentication field.
func (d *Document) AddPublicKey(pk *DocPublicKey, addRefToAuth bool, addFragment bool) error {
	// If ID is not given on public key, then add the doc owner ID by default and
	// add the next available key fragment value.
	// Overrides addFragment bool. if specific ID is needed, set ID/fragment before adding.
	if pk.ID == nil {
		pk.ID = CopyDID(&d.ID)
		pk.SetIDFragment(d.NextKeyFragment())

	} else {
		if !addFragment && pk.ID.Fragment == "" {
			return errors.Errorf("no key id fragment found: %v", pk.ID.String())

		} else if addFragment {
			// Increment the standard "keys-"
			pk.SetIDFragment(d.NextKeyFragment())
		}
	}

	// If controller is not set, set it to the doc owner ID by default. If you want
	// something else, needs to be passed in.
	if pk.Controller == nil {
		pk.Controller = CopyDID(&d.ID)
	}

	// If pk already exists, return
	if PublicKeyInSlice(*pk, d.PublicKeys) {
		log.Infof("Public key is already in document: %+v", *pk)
		return nil
	}

	// Add new key to end of the list of keys
	d.PublicKeys = append(d.PublicKeys, *pk)

	if addRefToAuth {
		auth := DocAuthenicationWrapper{
			DocPublicKey: *pk,
			IDOnly:       true,
		}
		d.Authentications = append(d.Authentications, auth)
	}

	updated := time.Now().UTC()
	d.Updated = &updated

	return nil
}

// AddAuthentication adds another authentication value to the list.  Could be
// just a reference to an existing key or a key only used for authentication
func (d *Document) AddAuthentication(auth *DocAuthenicationWrapper, addFragment bool) error {
	// If key is not given on public key,
	// then add it with the next available key fragment
	// Overrides addFragment bool. if specific ID is needed, set ID/fragment before adding.
	if auth.ID == nil {
		auth.ID = CopyDID(&d.ID)
		auth.SetIDFragment(d.NextKeyFragment())

	} else {
		if !addFragment && auth.ID.Fragment == "" {
			return errors.Errorf("no auth key id fragment found: %v", auth.ID.String())
		} else if addFragment {
			// Increment the standard "keys-"
			auth.SetIDFragment(d.NextKeyFragment())
		}
	}

	if !auth.IDOnly {
		// If controller is not set if auth is key, set it to the doc owner ID
		// by default. If you want something else, needs to be passed in.
		if auth.Controller == nil {
			auth.Controller = CopyDID(&d.ID)
		}
	}

	// If auth already exists, return
	if AuthInSlice(*auth, d.Authentications) {
		log.Infof("Auth is already in document: %+v", *auth)
		return nil
	}

	if auth.IDOnly {
		found := false
		// Ensure our public key exists for this reference
		for _, k := range d.PublicKeys {
			if auth.ID.String() == k.ID.String() {
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("auth ref reference has no matching key: %v", auth.ID.String())
		}
	}

	// Add authentication to the list of authentications
	d.Authentications = append(d.Authentications, *auth)

	updated := time.Now().UTC()
	d.Updated = &updated

	return nil
}

// AddService adds another service value to the doc.
func (d *Document) AddService(srv *DocService) error {
	if srv.ID.String() == "" {
		return errors.Errorf("no service id found, required")
	}

	if srv.ID.Fragment == "" {
		return errors.Errorf("no service fragment found, required: %v", srv.ID.String())
	}

	// If service already exists, return
	if ServiceInSlice(*srv, d.Services) {
		log.Infof("Service is already in document: %+v", *srv)
		return nil
	}

	// Add service to the list of services
	d.Services = append(d.Services, *srv)

	updated := time.Now().UTC()
	d.Updated = &updated

	return nil
}

// NextKeyFragment looks at all the authentication and public keys and
// finds the next incremented key fragment to use when adding a publicKey or
// authentication. Only supports key fragments (#key-*) right now.
func (d *Document) NextKeyFragment() string {
	keys := []string{}
	keyPrefix := "keys-"

	// Look at all the possible keys
	for _, k := range d.PublicKeys {
		if strings.HasPrefix(k.ID.Fragment, keyPrefix) {
			keys = append(keys, k.ID.Fragment)
		}
	}
	for _, k := range d.Authentications {
		if !k.IDOnly && strings.HasPrefix(k.ID.Fragment, keyPrefix) {
			keys = append(keys, k.ID.Fragment)
		}
	}

	if len(keys) == 0 {
		// keys-1
		return fmt.Sprintf("%v1", keyPrefix)
	}

	// Sort to get the max value
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	// Pull off value from last key and increment it
	recentKey := keys[0]
	spl := strings.SplitAfter(recentKey, keyPrefix)
	keyIncr := spl[1]

	incrInt, _ := strconv.Atoi(keyIncr)
	incrInt++

	return fmt.Sprintf("%v%v", keyPrefix, incrInt)
}

// DocPublicKey defines a publickey within a DID document
type DocPublicKey struct {
	ID                 *didlib.DID `json:"id"`
	Type               LDSuiteType `json:"type"`
	Controller         *didlib.DID `json:"controller"`
	PublicKeyPem       *string     `json:"publicKeyPem,omitempty"`
	PublicKeyJwk       *string     `json:"publicKeyJwk,omitempty"`
	PublicKeyHex       *string     `json:"publicKeyHex,omitempty"`
	PublicKeyBase64    *string     `json:"publicKeyBase64,omitempty"`
	PublicKeyBase58    *string     `json:"publicKeyBase58,omitempty"`
	PublicKeyMultibase *string     `json:"publicKeyMultibase,omitempty"`
	EthereumAddress    *string     `json:"ethereumAddress,omitempty"`
}

// SetIDFragment sets the ID fragment of the public key.  For convenience,
// returns the DocPublicKey for inline-ing.
func (p *DocPublicKey) SetIDFragment(fragment string) *DocPublicKey {
	// if p.ID == nil {
	// 	log.Errorf("public key ID is nil, not setting ID fragment")
	// 	return p
	// }
	p.ID.Fragment = fragment
	return p
}

// UnmarshalJSON implements the Unmarshaler interface for DocPublicKey
func (p *DocPublicKey) UnmarshalJSON(b []byte) error {
	type pkAlias DocPublicKey
	aux := &struct {
		ID         string `json:"id"`
		Controller string `json:"controller"`
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
	p.ID = id

	if aux.Controller != "" {
		controller, err := didlib.Parse(aux.Controller)
		if err != nil {
			return errors.Wrap(err, "unable to parse did for public key")
		}
		p.Controller = controller
	}

	if p.EthereumAddress != nil && *p.EthereumAddress != "" {
		p.EthereumAddress = utils.StrToPtr(eth.NormalizeEthAddress(*p.EthereumAddress))
	}

	return nil
}

// MarshalJSON implements the Marshaler interface for DocPublicKey
func (p *DocPublicKey) MarshalJSON() ([]byte, error) {
	type pkAlias DocPublicKey
	aux := &struct {
		ID         string `json:"id"`
		Controller string `json:"controller"`
		*pkAlias
	}{
		ID:      p.ID.String(),
		pkAlias: (*pkAlias)(p),
	}

	if p.Controller != nil && p.Controller.String() != "" {
		aux.Controller = p.Controller.String()
	}

	return json.Marshal(aux)
}

// DocAuthenicationWrapper allows us to handle two different types for an authentication
// value.  This can either be an ID to a public key or a public key.
type DocAuthenicationWrapper struct {
	DocPublicKey
	IDOnly bool `json:"-"`
}

// SetIDFragment sets the ID fragment of the authentication.  For convenience,
// returns the DocAuthenticationWrapper for inline-ing.
func (a *DocAuthenicationWrapper) SetIDFragment(fragment string) *DocAuthenicationWrapper {
	a.ID.Fragment = fragment
	return a
}

// UnmarshalJSON implements the Unmarshaler interface for DocAuthenticationWrapper
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

	a.ID = d
	a.IDOnly = true

	return nil
}

// MarshalJSON implements the Marshaler interface for DocAuthenticationWrapper
func (a *DocAuthenicationWrapper) MarshalJSON() ([]byte, error) {
	if a.IDOnly {
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
	ID          didlib.DID `json:"id"`
	Type        string     `json:"type"`
	Description string     `json:"description,omitempty"`
	PublicKey   string     `json:"publicKey,omitempty"`
	// DocServiceEndpoint could be a JSON-LD object or a URI
	// https://github.com/piprate/json-gold
	// string or map[string]interface{}
	ServiceEndpoint interface{} `json:"serviceEndpoint"`

	// DocServiceEndpoint values stored here as the correct type
	// Use these to access the values for DocServiceEndpoint
	ServiceEndpointURI *string                `json:"-"`
	ServiceEndpointLD  map[string]interface{} `json:"-"`
}

// PopulateServiceEndpointVals populate the ServiceEndpointURI or ServiceEdnpointLD
// based on the ServiceEndpoint interface{} value.
func (s *DocService) PopulateServiceEndpointVals() error {
	// Validate types for service endpoint.  Can either be a string (URI)
	// or a JSON-LD object
	switch val := s.ServiceEndpoint.(type) {
	case string:
		// valid type for URIs
		s.ServiceEndpointURI = &val
	case map[string]interface{}:
		// valid type
		// TODO: do more for validation of JSON-LD
		s.ServiceEndpointLD = val
	default:
		return errors.Errorf("invalid type for service endpoint value: %T", val)
	}
	return nil
}

// UnmarshalJSON implements the Unmarshaler interface for DocService
func (s *DocService) UnmarshalJSON(b []byte) error {
	type alias DocService
	aux := &struct {
		ID string `json:"id"`
		*alias
	}{
		alias: (*alias)(s),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal public key")
	}

	id, err := didlib.Parse(aux.ID)
	if err != nil {
		return errors.Wrap(err, "unable to parse did for service")
	}
	s.ID = *id

	err = s.PopulateServiceEndpointVals()
	if err != nil {
		return errors.Wrap(err, "unable to populate endpoint values")
	}

	return nil
}

// MarshalJSON implements the Marshaler interface for DocService
func (s *DocService) MarshalJSON() ([]byte, error) {
	type alias DocService
	aux := &struct {
		ID string `json:"id"`
		*alias
	}{
		ID:    s.ID.String(),
		alias: (*alias)(s),
	}

	return json.Marshal(aux)
}
