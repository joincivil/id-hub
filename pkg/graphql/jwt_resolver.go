package graphql

import (
	"context"
	"fmt"
	"strconv"

	log "github.com/golang/glog"
	"github.com/joincivil/id-hub/pkg/auth"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	didlib "github.com/ockam-network/did"
	"github.com/pkg/errors"
)

// Edge returns an edge resolver
func (r *Resolver) Edge() EdgeResolver {
	return &edgeResolver{r}
}

// FindEdges returns all edges for a did
func (r *queryResolver) FindEdges(ctx context.Context, in *FindEdgesInput) ([]*claimsstore.JWTClaimPostgres, error) {
	if in.FromDid == nil {
		return nil, errors.New("currently only supports searching by fromdid")
	}
	d, err := didlib.Parse(*in.FromDid)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing did in claim get")
	}

	tokens, err := r.JWTService.GetJWTSforDID(d)
	if err != nil {
		return nil, errors.Wrap(err, "error getting tokens for did")
	}

	claims := []*claimsstore.JWTClaimPostgres{}

	for _, v := range tokens {
		claim, err := claimsstore.TokenToJWTClaimPostgres(v)
		if err != nil {
			return claims, errors.Wrap(err, "FindEdges could not convert token to model")
		}
		claims = append(claims, claim)
	}

	return claims, nil
}

// AddEdge add a new edge
func (r *mutationResolver) AddEdge(ctx context.Context, edgeJwt *string) (*claimsstore.JWTClaimPostgres, error) {
	// Auth needed here, DID owner only
	fcd, authErr := auth.ForContext(ctx, r.DidService, nil)
	if authErr != nil {
		log.Infof("Access denied err: %v", authErr)
		return nil, ErrAccessDenied
	}

	senderDID, err := didlib.Parse(fcd.Did)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse sender did")
	}

	token, err := r.JWTService.AddJWTClaim(*edgeJwt, senderDID)
	if err != nil {
		return nil, errors.Wrap(err, "AddEdge couldn't add jwt")
	}

	return claimsstore.TokenToJWTClaimPostgres(token)
}

type edgeResolver struct{ *Resolver }

// From resolves issuer to from
func (r *edgeResolver) From(ctx context.Context, obj *claimsstore.JWTClaimPostgres) (string, error) {
	return obj.Issuer, nil
}

// To resolves subject to to
func (r *edgeResolver) To(ctx context.Context, obj *claimsstore.JWTClaimPostgres) (*string, error) {
	return &obj.Subject, nil
}

// Time resolves issuedat to time
func (r *edgeResolver) Time(ctx context.Context, obj *claimsstore.JWTClaimPostgres) (string, error) {
	return strconv.FormatInt(obj.IssuedAt, 10), nil
}

// Proof gets a proof for the edge
func (r *edgeResolver) Proof(ctx context.Context, obj *claimsstore.JWTClaimPostgres) ([]Proof, error) {
	proof, err := r.JWTService.GenerateProof(obj.JWT)

	if err != nil {
		return nil, errors.Wrap(err, "edge resolver couldnt generate the proof")
	}

	rootProof := RootOnBlockChainProof{
		Type:             claimtypes.RootInContract,
		BlockNumber:      fmt.Sprintf("%d", proof.BlockNumber),
		Root:             proof.Root.Hex(),
		ContractAddress:  proof.ContractAddress.Hex(),
		CommitterAddress: proof.CommitterAddress.Hex(),
		TxHash:           proof.TXHash.Hex(),
	}

	inTreeProof := ClaimRegisteredProof{
		Type:                   claimtypes.MerkleProof,
		Did:                    proof.DID,
		ExistsInDIDMTProof:     proof.ExistsInDIDMTProof,
		NotRevokedInDIDMTProof: proof.NotRevokedInDIDMTProof,
		DidMTRootExistsProof:   proof.DIDRootExistsProof,
		DidRootExistsVersion:   int(proof.DIDRootExistsVersion),
		Root:                   proof.Root.Hex(),
		DidMTRoot:              proof.DIDRoot.Hex(),
	}

	return []Proof{inTreeProof, rootProof}, nil
}
