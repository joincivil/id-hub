package auth

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	log "github.com/golang/glog"
	didlib "github.com/ockam-network/did"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	ceth "github.com/joincivil/go-common/pkg/eth"
	ctime "github.com/joincivil/go-common/pkg/time"

	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

const (
	defaultGracePeriod = 60 * 5 // 5 mins
)

// VerifyEcdsaRequestSignatureWithDid checks the did document for keys and
// verifies the signatures using the dids ECDSA public keys
func VerifyEcdsaRequestSignatureWithDid(ds *did.Service, keyType linkeddata.SuiteType,
	signature string, ts int, didStr string) error {
	if !linkeddata.IsEcdsaKeySuiteType(keyType) {
		return errors.New("supports ecdsa only")
	}

	doc, err := ds.GetDocument(didStr)
	if err != nil {
		return errors.Wrapf(err, "did not found for %v", didStr)
	}
	if doc == nil {
		return errors.Errorf("did doc not found for %v", didStr)
	}

	return VerifyEcdsaRequestSignatureWithPks(doc.PublicKeys, keyType, signature, ts, didStr)
}

// VerifyEcdsaRequestSignatureWithPks checks a slice of public keys and verifies
// the signature against keys of key suite type ECDSA. didStr only affects the
// signed request message value and can be omitted (look at RequestMessage for more details).
func VerifyEcdsaRequestSignatureWithPks(pks []did.DocPublicKey, keyType linkeddata.SuiteType,
	signature string, ts int, didStr string) error {
	if !linkeddata.IsEcdsaKeySuiteType(keyType) {
		return errors.New("supports ecdsa only")
	}

	if len(pks) == 0 {
		return errors.New("no publickeys found")
	}

	var err error
	var retErr error
	var pubKey *string
	var valid bool
	verified := false

KeyLoop:
	for _, key := range pks {
		if key.Type == keyType {
			pubKey, err = did.KeyFromType(&key)
			if err != nil {
				log.Errorf("Error getting key from type: err: %v", err)
				retErr = err
				continue
			}

			valid, err = VerifyEcdsaRequestSignature(*pubKey, signature, didStr, ts)
			if err != nil {
				log.Errorf("Error verifying signature: err: %v", err)
				retErr = err
			}

			if valid {
				verified = true
				break KeyLoop
			}
		}
	}

	if retErr != nil {
		return errors.Wrap(retErr, "error when verifying signature")
	}

	if !verified {
		return errors.New("signature is invalid")
	}
	return nil
}

// VerifyEcdsaRequestSignature determines if a signature is valid given the ECDSA public key
// and a message derived from a message containing a did and the request timestamp.
// NOTE: The did is only validated for correctness, but has not validated to see if
// there is a corresponding did document.  That should occur before this method is called.
// The message to be verified is "<did> request @ <timestamp>"
func VerifyEcdsaRequestSignature(pubKey string, signature string,
	did string, reqTs int) (bool, error) {
	if did != "" {
		_, err := didlib.Parse(did)
		if err != nil {
			return false, errors.Wrap(err, "error parsing did for signature")
		}
	}

	// Signed message should be did and timestamp related
	msg := RequestMessage(did, reqTs)

	pubKeyBys, err := hex.DecodeString(pubKey)
	if err != nil {
		return false, errors.Wrap(err, "error decoding pubkey hex")
	}
	pk, err := crypto.UnmarshalPubkey(pubKeyBys)
	if err != nil {
		return false, errors.Wrap(err, "error unmarshalling to public key")
	}

	// Verify that the signature is correct
	verified, err := ceth.VerifyEthSignatureWithPubkey(*pk, msg, signature)
	if err != nil {
		return false, errors.Wrap(err, "error verifying signature")
	}

	if !verified {
		return false, nil
	}

	// Check timestamp to ensure no long term replays
	tsDiff := ctime.CurrentEpochSecsInInt() - reqTs

	// There is something weird if reqTs is greater than current time
	if tsDiff < 0 {
		log.Infof("Request timestamp is greater than current time: %v", reqTs)
		return false, nil
	}
	// Grace period for validity has expired
	if tsDiff > defaultGracePeriod {
		return false, nil
	}

	return true, nil
}

// SignEcdsaRequestMessage is a convenience function to sign a message used for
// API requests
func SignEcdsaRequestMessage(privKey *ecdsa.PrivateKey, did string, reqTs int) (string, error) {
	if did != "" {
		_, err := didlib.Parse(did)
		if err != nil {
			return "", errors.Wrap(err, "invalid did for signing")
		}
	}

	signature, err := ceth.SignEthMessage(privKey, RequestMessage(did, reqTs))
	if err != nil {
		return "", errors.Wrap(err, "error signing request message")
	}
	return signature, nil
}

// RequestMessage returns the default message to be signed for API
// requests
func RequestMessage(did string, reqTs int) string {
	if did == "" {
		return fmt.Sprintf("request @ %v", reqTs)
	}
	return fmt.Sprintf("%v request @ %v", did, reqTs)
}