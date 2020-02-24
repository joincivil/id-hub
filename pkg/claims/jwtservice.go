package claims

import (
	"encoding/hex"

	"github.com/dgrijalva/jwt-go"
	log "github.com/golang/glog"
	icore "github.com/iden3/go-iden3-core/core"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/didjwt"
	"github.com/joincivil/id-hub/pkg/pubsub"
	"github.com/joincivil/id-hub/pkg/utils"
	didlib "github.com/ockam-network/did"
	"github.com/pkg/errors"
)

// JWTService is a service for registering JWT claims
type JWTService struct {
	didJWTService *didjwt.Service
	jwtPersister  *claimsstore.JWTClaimPGPersister
	claimService  *Service
	natsService   pubsub.Interface
}

// NewJWTService creates a new instance of the service
func NewJWTService(didJWTService *didjwt.Service,
	jwtPersister *claimsstore.JWTClaimPGPersister,
	claimService *Service, natsService pubsub.Interface) *JWTService {
	return &JWTService{
		didJWTService: didJWTService,
		jwtPersister:  jwtPersister,
		claimService:  claimService,
		natsService:   natsService,
	}
}

// AddJWTClaim adds a new jwt claim to it's issuers tree
func (s *JWTService) AddJWTClaim(tokenString string, senderDID *didlib.DID) (*jwt.Token, error) {
	token, hash, err := s.jwtPersister.AddJWT(tokenString, senderDID)
	if err != nil {
		return nil, errors.Wrap(err, "AddJWTClaim error adding JWT to db")
	}

	issuer, err := getIssuerDIDfromToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "AddJWTClaim error parsing issuer did")
	}

	didMT, err := s.claimService.BuildDIDMt(issuer)
	if err != nil {
		return nil, errors.Wrap(err, "AddJWTClaim error building didmt")
	}

	hashb, err := hex.DecodeString(hash)
	if err != nil {
		return nil, errors.Wrap(err, "claimcontent.decodestring")
	}
	if len(hashb) > 34 {
		return nil, errors.New("hash hex string is the wrong size")
	}
	hashb34 := [34]byte{}
	copy(hashb34[:], hashb)

	claim, err := claimtypes.NewClaimRegisteredDocument(hashb34, issuer, claimtypes.JWTDocType)
	if err != nil {
		return nil, errors.Wrap(err, "AddJWTClaim error creating registered document claim")
	}

	err = didMT.Add(claim.Entry())
	if err != nil {
		return nil, errors.Wrap(err, "AddJWTClaim add claim to did mt")
	}

	err = s.claimService.AddNewRootClaim(issuer)
	if err != nil {
		return nil, errors.Wrap(err, "AddJWTClaim.addnewrootclaim")
	}

	err = s.natsService.PublishAdd(token)
	if err != nil {
		return nil, errors.Wrap(err, "AddJWTClaim couldn't publish to nats")
	}

	return token, nil
}

func (s *JWTService) makeRegisteredDocClaimFromJWT(token string,
	issuer *didlib.DID) (*claimtypes.ClaimRegisteredDocument, error) {
	hash, err := utils.CreateMultihash([]byte(token))
	if err != nil {
		return nil, errors.Wrap(err, "makeRegisteredDocClaimFromJWT error creating multihash")
	}
	hash34 := [34]byte{}
	copy(hash34[:], hash)
	return claimtypes.NewClaimRegisteredDocument(hash34, issuer, claimtypes.JWTDocType)
}

// RevokeJWTClaim takes a token and revokes it in the merkle tree
func (s *JWTService) RevokeJWTClaim(tokenString string) error {
	token, err := s.didJWTService.ParseJWT(tokenString)
	if err != nil {
		return errors.Wrap(err, "RevokeJWTClaim couldn't parse token")
	}

	issuer, err := getIssuerDIDfromToken(token)
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

	err = s.natsService.PublishRevoke(token)
	if err != nil {
		return errors.Wrap(err, "RevokeJWTClaim couldn't publish to nats")
	}

	return nil
}

// GetJWTSforDID returns all jwt claims for a DID
func (s *JWTService) GetJWTSforDID(userDID *didlib.DID) ([]*jwt.Token, error) {
	claims, err := s.claimService.GetMerkleTreeClaimsForDid(userDID)
	if err != nil {
		return nil, err
	}
	tokens := make([]*jwt.Token, 0, len(claims))
	for _, v := range claims {
		switch tv := v.(type) {
		case claimtypes.ClaimRegisteredDocument, *claimtypes.ClaimRegisteredDocument:
			var regDoc claimtypes.ClaimRegisteredDocument
			d, ok := tv.(*claimtypes.ClaimRegisteredDocument)
			if ok {
				regDoc = *d
			} else {
				regDoc = tv.(claimtypes.ClaimRegisteredDocument)
			}

			if regDoc.DocType == claimtypes.JWTDocType {
				claimHash := hex.EncodeToString(regDoc.ContentHash[:])

				token, err := s.jwtPersister.GetJWTByMultihash(claimHash)
				if err != nil {
					return nil, errors.Wrap(err, "GetJWTSforDID error fetching token from db")
				}
				tokens = append(tokens, token)
			}
		case *icore.ClaimAuthorizeKSignSecp256k1:
			// Known claim type to ignore here
		default:
			log.Errorf("Unknown claim type, is %T", v)
		}
	}
	return tokens, nil
}

// GenerateProof creates a proof from a jwt
func (s *JWTService) GenerateProof(tokenString string) (*MTProof, error) {
	token, err := s.didJWTService.ParseJWT(tokenString)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof couldn't parse token")
	}

	issuer, err := getIssuerDIDfromToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof error parsing issuer did")
	}

	regDocClaim, err := s.makeRegisteredDocClaimFromJWT(tokenString, issuer)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateProof couldn't make reg doc claim")
	}

	return s.claimService.GenerateProofRegistedDocument(regDocClaim, issuer)
}
