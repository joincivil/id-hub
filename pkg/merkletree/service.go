package merkletree

import (
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/didjwt"
	"github.com/joincivil/id-hub/pkg/utils"
	didlib "github.com/ockam-network/did"
	"github.com/pkg/errors"
)

// Service is a service for registering JWT claims
type Service struct {
	didJWTService *didjwt.Service
	claimService  *claims.Service
}

// NewService creates a new instance of the service
func NewService(didJWTService *didjwt.Service,
	claimService *claims.Service) *Service {
	return &Service{
		didJWTService: didJWTService,
		claimService:  claimService,
	}
}

func (s *Service) getIssuer(tokenString string) (*didlib.DID, error) {
	token, err := s.didJWTService.ParseJWT(tokenString)
	if err != nil {
		return nil, errors.Wrap(err, "AddEntry failed to parse token")
	}

	return claims.GetIssuerDIDfromToken(token)
}

// AddEntry adds a new jwt claim to it's issuers tree
func (s *Service) AddEntry(tokenString string, sender string) (string, error) {
	issuer, err := s.getIssuer(tokenString)
	claimtype := claimtypes.JWTDocType
	if err != nil {
		issuer, err = didlib.Parse(sender)
		claimtype = claimtypes.RawDataDocType
		if err != nil {
			return "", errors.Wrap(err, "AddEntry unable to parse did from token or sender")
		}
	}

	didMT, err := s.claimService.BuildDIDMt(issuer)
	if err != nil {
		return "", errors.Wrap(err, "AddEntry error building didmt")
	}

	hash, err := utils.CreateMultihash([]byte(tokenString))
	if err != nil {
		return "", errors.Wrap(err, "AddEntry couldn't create hash of token")
	}

	if len(hash) > 34 {
		return "", errors.New("hash hex string is the wrong size")
	}
	hashb34 := [34]byte{}
	copy(hashb34[:], hash)

	claim, err := claimtypes.NewClaimRegisteredDocument(hashb34, issuer, claimtype)
	if err != nil {
		return "", errors.Wrap(err, "AddEntry error creating registered document claim")
	}

	err = didMT.Add(claim.Entry())
	if err != nil {
		return "", errors.Wrap(err, "AddEntry add claim to did mt")
	}

	err = s.claimService.AddNewRootClaim(issuer)
	if err != nil {
		return "", errors.Wrap(err, "AddEntry.addnewrootclaim")
	}

	return tokenString, nil
}

func (s *Service) makeRegisteredDocClaimFromJWT(token string,
	issuer *didlib.DID) (*claimtypes.ClaimRegisteredDocument, error) {
	hash, err := utils.CreateMultihash([]byte(token))
	if err != nil {
		return nil, errors.Wrap(err, "makeRegisteredDocClaimFromJWT error creating multihash")
	}
	hash34 := [34]byte{}
	copy(hash34[:], hash)
	return claimtypes.NewClaimRegisteredDocument(hash34, issuer, claimtypes.JWTDocType)
}

// RevokeEntry takes a token and revokes it in the merkle tree
func (s *Service) RevokeEntry(tokenString string) error {
	token, err := s.didJWTService.ParseJWT(tokenString)
	if err != nil {
		return errors.Wrap(err, "RevokeJWTClaim couldn't parse token")
	}

	issuer, err := claims.GetIssuerDIDfromToken(token)
	if err != nil {
		return errors.Wrap(err, "RevokeJWTClaim error parsing issuer did")
	}

	didMt, err := s.claimService.BuildDIDMt(issuer)
	if err != nil {
		return errors.Wrap(err, "RevokeJWTClaim.builddidMt")
	}

	regDocClaim, err := s.makeRegisteredDocClaimFromJWT(tokenString, issuer)
	if err != nil {
		return errors.Wrap(err, "RevokeJWTClaim couldn't make reg doc claim")
	}

	regDocClaim.Version = 1

	err = didMt.Add(regDocClaim.Entry())
	if err != nil {
		return errors.Wrap(err, "RevokeJWTClaim.add")
	}

	err = s.claimService.AddNewRootClaim(issuer)
	if err != nil {
		return errors.Wrap(err, "RevokeJWTClaim.addnewrootclaim")
	}

	return nil

}

// GenerateProof creates a proof from a jwt
func (s *Service) GenerateProof(tokenString string) (*claims.MTProof, error) {
	token, err := s.didJWTService.ParseJWT(tokenString)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof couldn't parse token")
	}

	issuer, err := claims.GetIssuerDIDfromToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof error parsing issuer did")
	}

	regDocClaim, err := s.makeRegisteredDocClaimFromJWT(tokenString, issuer)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof couldn't make reg doc claim")
	}

	return s.claimService.GenerateProofRegistedDocument(regDocClaim, issuer)
}
