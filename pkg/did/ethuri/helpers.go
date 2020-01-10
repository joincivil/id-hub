package ethuri

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joincivil/id-hub/pkg/did"
	didlib "github.com/ockam-network/did"
	"github.com/pkg/errors"
)

// GenerateNewDocument generates a new DID and DID document and sets the first public
// key for the DID. Sets the public key into the publicKeys field and adds a
// reference to the key to the authentication field. If firstPK has an empty ID
// field, will populate it with the new DID.
func GenerateNewDocument(firstPK *did.DocPublicKey, addRefToAuth bool,
	addFragment bool) (*did.Document, error) {

	newDID, err := GenerateEthURIDID()
	if err != nil {
		return nil, errors.Wrap(err, "error generating new ethuri did")
	}

	// If firstPK ID is not set, then set it to the newly created DID
	if firstPK.ID == nil {
		firstPK.ID = newDID
		firstPK.Controller = did.CopyDID(newDID)
	}

	if !did.ValidDocPublicKey(firstPK) {
		return nil, errors.New("invalid doc public key")
	}

	doc, err := InitializeNewDocument(newDID, firstPK, addRefToAuth, addFragment)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing new did document")
	}

	return doc, nil
}

// GenerateEthURIDID generates a new EthURI method DID
func GenerateEthURIDID() (*didlib.DID, error) {
	// Generate a new UUID v4
	newUUID := uuid.New()
	didStr := fmt.Sprintf("%s:%s", EthURISchemeMethod, newUUID.String())
	return didlib.Parse(didStr)
}

// InitializeNewDocument generates a simple version of a DID document given
// the DID and an initial public key.
func InitializeNewDocument(d *didlib.DID, firstPK *did.DocPublicKey, addRefToAuth bool,
	addFragment bool) (*did.Document, error) {
	if !did.ValidDocPublicKey(firstPK) {
		return nil, errors.New("invalid doc public key")
	}

	created := time.Now().UTC()
	updated := time.Now().UTC()

	doc := &did.Document{
		Context:         did.DefaultDIDContextV1,
		ID:              *d,
		Controller:      d,
		PublicKeys:      []did.DocPublicKey{},
		Authentications: []did.DocAuthenicationWrapper{},
		Services:        []did.DocService{},
		Created:         &created,
		Updated:         &updated,
	}

	err := doc.AddPublicKey(firstPK, addRefToAuth, addFragment)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
