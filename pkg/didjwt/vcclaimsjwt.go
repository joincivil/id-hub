package didjwt

import "github.com/dgrijalva/jwt-go"

// VCClaimsJWT is a jwt claims struct with a data field for vc data
type VCClaimsJWT struct {
	Data string `json:"data"`
	jwt.StandardClaims
}
