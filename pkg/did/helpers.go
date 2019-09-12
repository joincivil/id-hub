package did

import (
	"encoding/hex"
	"fmt"
	"time"

	log "github.com/golang/glog"
	"github.com/google/uuid"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/crypto"
	didlib "github.com/ockam-network/did"
)

// GenerateNewDocument generates a new DID and DID document and sets the first public
// key for the DID. Sets the public key into the publicKeys field and adds a
// reference to the key to the authentication field. If firstPK has an empty ID
// field, will populate it with the new DID.
func GenerateNewDocument(firstPK *DocPublicKey) (*Document, error) {
	if !ValidDocPublicKey(firstPK) {
		return nil, errors.New("invalid doc public key")
	}

	newDID, err := GenerateEthURIDID()
	if err != nil {
		return nil, errors.Wrap(err, "error generating new ethuri did")
	}

	// If firstPK ID is not set, then set it to the newly created DID
	if firstPK.ID == nil {
		firstPK.ID = newDID
		firstPK.Controller = CopyDID(newDID)
	}

	doc, err := InitializeNewDocument(newDID, firstPK)
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
func InitializeNewDocument(did *didlib.DID, firstPK *DocPublicKey) (*Document, error) {
	if !ValidDocPublicKey(firstPK) {
		return nil, errors.New("invalid doc public key")
	}

	created := time.Now().UTC()
	updated := time.Now().UTC()

	doc := &Document{
		Context:         DefaultDIDContextV1,
		ID:              *did,
		Controller:      did,
		PublicKeys:      []DocPublicKey{},
		Authentications: []DocAuthenicationWrapper{},
		Created:         &created,
		Updated:         &updated,
	}

	err := doc.AddPublicKey(firstPK.SetIDFragment(doc.NextKeyFragment()), true)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// CopyDID is a convenience function to make a copy a DID struct
func CopyDID(d *didlib.DID) *didlib.DID {
	cpy, err := didlib.Parse(d.String())
	if err != nil {
		log.Errorf("Error parsing did string for copy: err: %v", err)
	}
	return cpy
}

// ValidDid returns true if the given did string is of a valid DID format
func ValidDid(did string) bool {
	_, err := didlib.Parse(did)
	return err == nil
}

// ValidateBuildDocPublicKey is a convenience function to validate the DocPublicKey
// and populate a new DocPublicKey with the type and key value. Returns a pre-populated
// DocPublicKey with the correct type and PublicKey* field populated for that type.
// ID and other fields are not set.
func ValidateBuildDocPublicKey(keyType LDSuiteType, keyValue string) *DocPublicKey {
	pk := &DocPublicKey{
		Type:         keyType,
		PublicKeyHex: keyValue,
	}
	if !ValidDocPublicKey(pk) {
		return nil
	}
	return pk
}

// ValidDocPublicKey ensures that the given DocPublicKey is of a supported type,
// has a valid key for that type and is using the correct public key field.
// Returns true if it is valid, false if not.
func ValidDocPublicKey(pk *DocPublicKey) bool {
	// Supports only Secp256k1 hex keys for now
	switch pk.Type {
	case LDSuiteTypeSecp256k1Verification:
		if pk.PublicKeyHex == "" {
			log.Errorf("publicKeyHex is not populated for SECP256k1")
			return false
		}
		// Hex keys do not have 0x prefix
		bys, err := hex.DecodeString(pk.PublicKeyHex)
		if err != nil {
			log.Errorf("unable to decode pub key hex str for SECP256k1: err: %v", err)
			return false
		}
		// Try to unmarshal to ensure valid key
		_, err = crypto.UnmarshalPubkey(bys)
		if err != nil {
			log.Errorf("invalid pub key value for SECP256k1: err: %v", err)
			return false
		}

		return true
	}
	return false
}
