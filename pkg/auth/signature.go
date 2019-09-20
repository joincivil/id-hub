package auth

import (
	"crypto/ecdsa"
	"fmt"

	log "github.com/golang/glog"
	didlib "github.com/ockam-network/did"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	ceth "github.com/joincivil/go-common/pkg/eth"
	ctime "github.com/joincivil/go-common/pkg/time"
)

// VerifyEcdsaRequestSignature determines if a signature is valid given the public key
// and a message derived from a message containing a did and the request timestamp.
// NOTE: The did is only validated for correctness, but has not validated to see if
// there is a corresponding did document.  That should occur before this method is called.
// The message to be verified is "<did> request @ <timestamp>"
func VerifyEcdsaRequestSignature(pubKey string, signature string,
	did string, reqTs int) (bool, error) {
	_, err := didlib.Parse(did)
	if err != nil {
		return false, errors.Wrap(err, "error parsing did for signature")
	}

	// Signed message should be did and timestamp related
	msg := RequestMessage(did, reqTs)
	pk, err := crypto.UnmarshalPubkey([]byte(pubKey))
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
	_, err := didlib.Parse(did)
	if err != nil {
		return "", errors.Wrap(err, "invalid did for signing")
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
	return fmt.Sprintf("%v request @ %v", did, reqTs)
}
