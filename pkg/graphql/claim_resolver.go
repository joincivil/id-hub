package graphql

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/golang/glog"
	didlib "github.com/ockam-network/did"

	"github.com/pkg/errors"

	"github.com/joincivil/go-common/pkg/article"
	"github.com/joincivil/id-hub/pkg/auth"
	"github.com/joincivil/id-hub/pkg/claimtypes"
	"github.com/joincivil/id-hub/pkg/utils"
)

// ResolverRoot

// ArticleMetadata returns the resolver for article metadata
func (r *Resolver) ArticleMetadata() ArticleMetadataResolver {
	return &articleMetadataResolver{r}
}

// Claim returns the resolver for claims
func (r *Resolver) Claim() ClaimResolver {
	return &claimResolver{r}
}

// Queries

func (r *queryResolver) ClaimGet(ctx context.Context, in *ClaimGetRequestInput) (
	*ClaimGetResponse, error) {
	d, err := didlib.Parse(in.Did)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing did in claim get")
	}
	clms, err := r.ClaimService.GetMerkleTreeClaimsForDid(d)
	if err != nil {
		return nil, errors.Wrap(err, "error getting claims in claim get")
	}

	creds, err := r.ClaimService.ClaimsToContentCredentials(clms)
	if err != nil {
		return nil, errors.Wrap(err, "error converting claims to creds")
	}

	return &ClaimGetResponse{Claims: creds}, nil
}

func (r *queryResolver) ClaimProof(ctx context.Context, in *ClaimProofRequestInput) (
	*ClaimProofResponse, error) {
	claimSaveInput := &ClaimSaveRequestInput{
		Claim:     in.Claim,
		ClaimJSON: in.ClaimJSON,
	}
	cc, err := InputClaimToContentCredential(claimSaveInput)
	if err != nil {
		return nil, errors.Wrap(err, "error converting claim to credential")
	}

	requesterDid, err := didlib.Parse(in.Did)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing did")
	}

	proof, err := r.ClaimService.GenerateProof(cc, requesterDid)
	if err != nil {
		return nil, errors.Wrap(err, "error generating proof that claim is in tree")
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

	cc.Proof = []interface{}{cc.Proof, inTreeProof, rootProof}

	claimRaw, err := json.Marshal(cc)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't marshal json from claim")
	}

	return &ClaimProofResponse{
		Claim:    cc,
		ClaimRaw: string(claimRaw),
	}, nil
}

// Mutations

func (r *mutationResolver) ClaimSave(ctx context.Context, in *ClaimSaveRequestInput) (
	*ClaimSaveResponse, error) {
	var err error

	// Auth needed here, DID owner only
	fcd, authErr := auth.ForContext(ctx, r.DidService, nil)
	if authErr != nil {
		log.Infof("Access denied err: %v", authErr)
		return nil, ErrAccessDenied
	}

	cc, err := InputClaimToContentCredential(in)
	if err != nil {
		return nil, errors.Wrap(err, "error converting claim to credential")
	}

	// Auth check to ensure that the requestor auth did matches the issuer did
	// value in the claim.
	if cc.Issuer != fcd.Did {
		log.Infof("Access denied, requestor did does not match issuer did: %v, %v",
			cc.Issuer, fcd.Did)
		return nil, ErrAccessDenied
	}

	issuerDID, err := didlib.Parse(cc.Issuer)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing issuer did")
	}

	err = r.ClaimService.CreateTreeForDID(issuerDID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create tree for did if not exists")
	}

	err = r.ClaimService.ClaimContent(cc)
	if err != nil {
		return nil, errors.Wrap(err, "error calling claimcontent")
	}

	return &ClaimSaveResponse{Claim: cc}, nil
}

// Claim Resolvers

type articleMetadataResolver struct{ *Resolver }

func (r *articleMetadataResolver) RevisionDate(ctx context.Context, obj *article.Metadata) (*string, error) {
	opd := obj.RevisionDate.Format(timeFormat)
	return utils.StrToPtr(opd), nil
}
func (r *articleMetadataResolver) OriginalPublishDate(ctx context.Context, obj *article.Metadata) (*string, error) {
	opd := obj.OriginalPublishDate.Format(timeFormat)
	return utils.StrToPtr(opd), nil
}

type claimResolver struct{ *Resolver }

func (r *claimResolver) Type(ctx context.Context, obj *claimtypes.ContentCredential) ([]string, error) {
	ts := make([]string, len(obj.Type))
	for ind, val := range obj.Type {
		ts[ind] = string(val)
	}
	return ts, nil
}
func (r *claimResolver) IssuanceDate(ctx context.Context, obj *claimtypes.ContentCredential) (string, error) {
	return obj.IssuanceDate.Format(timeFormat), nil
}

func (r *claimResolver) CredentialSubject(ctx context.Context, obj *claimtypes.ContentCredential) (*ContentClaimCredentialSubject, error) {
	return &ContentClaimCredentialSubject{
		ID:       obj.CredentialSubject.ID,
		Metadata: &obj.CredentialSubject.Metadata,
	}, nil
}

func (r *claimResolver) Proof(ctx context.Context, obj *claimtypes.ContentCredential) ([]Proof, error) {
	switch val := obj.Proof.(type) {
	case []interface{}:
		proofs := make([]Proof, len(val))
		for i, v := range val {
			tv, ok := v.(Proof)
			if ok {
				proofs[i] = tv
			}
		}
		return proofs, nil

	case interface{}:
		return []Proof{val.(Proof)}, nil
	}
	return nil, errors.New("Invalid proof types")
}
