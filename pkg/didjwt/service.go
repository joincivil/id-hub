package didjwt

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
)

// Service is a service object for parsing didtype jwts
type Service struct {
	didService *did.Service
}

// NewService returns a new service
func NewService(didService *did.Service) *Service {
	return &Service{
		didService: didService,
	}
}

// ParseJWT takes a jwt token finds the correct key to verify it and returns the token
func (s *Service) ParseJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &VCClaimsJWT{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if claims, ok := token.Claims.(*VCClaimsJWT); ok && claims.Issuer != "" {
			didDoc, err := s.didService.GetDocument(claims.Issuer)
			if err != nil {
				return nil, err
			}
			var key *ecdsa.PublicKey
			for _, v := range didDoc.PublicKeys {
				// only support secp256r1 and secp256k1 for now
				if v.Type == linkeddata.SuiteTypeSecp256k1Verification ||
					v.Type == linkeddata.SuiteTypeSecp256r1Verification {
					key, err = v.AsEcdsaPubKey()
					if err != nil {
						continue
					}

					parts := strings.Split(token.Raw, ".")
					fParts := []string{parts[0], parts[1]}
					signingString := strings.Join(fParts, ".")
					err = jwt.SigningMethodES256.Verify(signingString, parts[2], key)
					if err == nil {
						break
					}
					key = nil
				}
			}
			if key == nil {
				return nil, errors.New("could not verify the token with any of the DID's keys")
			}
			return key, nil
		}
		return nil, errors.New("couldn't get public key from DID")
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
