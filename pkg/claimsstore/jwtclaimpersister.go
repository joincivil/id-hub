package claimsstore

import (
	"encoding/hex"

	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/didjwt"
	"github.com/multiformats/go-multihash"
	didlib "github.com/ockam-network/did"
	"github.com/pkg/errors"
)

// JWTClaimPostgres is a model for storing jwt type vcs
type JWTClaimPostgres struct {
	JWT    string `gorm:"not null"`
	Issuer string `gorm:"not null;index:jwtissuer"`
	Sender string `gorm:"not null;index:jwtsender"`
	Hash   string `gorm:"primary_key"`
}

// TableName sets the name of the table in the db
func (JWTClaimPostgres) TableName() string {
	return "jwt_claims"
}

func hashJWT(token string) (string, error) {
	hash := crypto.Keccak256([]byte(token))
	mHash, err := multihash.EncodeName(hash, "keccak-256")
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mHash), nil
}

// JWTClaimPGPersister is a postgres persister for JWT claims
type JWTClaimPGPersister struct {
	db            *gorm.DB
	didJWTService *didjwt.Service
}

// NewJWTClaimPGPersister returns a new JWTClaimPGPersister
func NewJWTClaimPGPersister(db *gorm.DB, didJWTService *didjwt.Service) *JWTClaimPGPersister {
	return &JWTClaimPGPersister{
		db:            db,
		didJWTService: didJWTService,
	}
}

// AddJWT adds a new jwt claim to the db
func (p *JWTClaimPGPersister) AddJWT(tokenString string, senderDID *didlib.DID) (*jwt.Token, string, error) {
	token, err := p.didJWTService.ParseJWT(tokenString)
	if err != nil {
		return nil, "", errors.Wrap(err, "addJWT failed to parse token")
	}
	hash, err := hashJWT(tokenString)
	if err != nil {
		return nil, "", errors.Wrap(err, "addJWT failed to hash token")
	}
	claims, ok := token.Claims.(*didjwt.VCClaimsJWT)
	if !ok {
		return nil, "", errors.New("invalid claims type")
	}

	claim := &JWTClaimPostgres{
		JWT:    tokenString,
		Issuer: claims.Issuer,
		Sender: senderDID.String(),
		Hash:   hash,
	}

	if err := p.db.Create(claim).Error; err != nil {
		return nil, "", errors.Wrap(err, "addJWT failed to save token to db")
	}

	return token, claim.Hash, nil
}

// GetJWTByHash returns a jwt from it's hash
func (p *JWTClaimPGPersister) GetJWTByHash(hash string) (*jwt.Token, error) {
	bytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, errors.Wrap(err, "GetJWTByHash failed to decode hash")
	}

	mHash, err := multihash.EncodeName(bytes, "keccak-256")
	if err != nil {
		return nil, errors.Wrap(err, "GetJWTByHash failed to create multihash")
	}

	mHashString := hex.EncodeToString(mHash)
	return p.GetJWTByMultihash(mHashString)
}

// GetJWTByMultihash returns a jwt from it's multihash
func (p *JWTClaimPGPersister) GetJWTByMultihash(mHash string) (*jwt.Token, error) {
	jwtClaim := &JWTClaimPostgres{}
	if err := p.db.Where(&JWTClaimPostgres{Hash: mHash}).First(jwtClaim).Error; err != nil {
		return nil, errors.Wrap(err, "GetJWTByMultihash failed to find claim")
	}
	return p.didJWTService.ParseJWT(jwtClaim.JWT)
}
