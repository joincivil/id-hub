package did

import (
	"fmt"

	log "github.com/golang/glog"

	didlib "github.com/ockam-network/did"
)

// MethodIDOnlyFromString returns did string without any fragments or paths from
// a full DID string
func MethodIDOnlyFromString(did string) (string, error) {
	d, err := didlib.Parse(did)
	if err != nil {
		return "", err
	}
	return MethodIDOnly(d), nil
}

// MethodIDOnly returns did string without any fragments or paths from a DID object
func MethodIDOnly(did *didlib.DID) string {
	return fmt.Sprintf("did:%v:%v", did.Method, did.ID)
}

// CopyDID is a convenience function to make a copy a DID struct
func CopyDID(d *didlib.DID) *didlib.DID {
	cpy, err := didlib.Parse(d.String())
	if err != nil {
		log.Errorf("Error parsing did string for copy: err: %v", err)
		return nil
	}
	return cpy
}

// ValidDid returns true if the given did string is of a valid DID format
func ValidDid(did string) bool {
	_, err := didlib.Parse(did)
	return err == nil
}
