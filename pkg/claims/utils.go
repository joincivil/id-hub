package claims

import (
	"encoding/json"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/didjwt"
	didlib "github.com/ockam-network/did"
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

// GetIssuerDIDfromToken returns the issuer did from a jwt
func GetIssuerDIDfromToken(token *jwt.Token) (*didlib.DID, error) {
	claims, ok := token.Claims.(*didjwt.VCClaimsJWT)

	if !ok {
		return nil, errors.New("invalids claims type on JWT")
	}

	issuer, err := didlib.Parse(claims.Issuer)
	if err != nil {
		return nil, errors.Wrap(err, "AddJWTClaim error parsing issuer did")
	}

	return issuer, nil
}
