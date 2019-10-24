package graphql

import (
	"context"

	log "github.com/golang/glog"
	didlib "github.com/ockam-network/did"

	"github.com/pkg/errors"

	"github.com/joincivil/go-common/pkg/article"
	"github.com/joincivil/id-hub/pkg/auth"
	"github.com/joincivil/id-hub/pkg/claimsstore"
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

func (r *claimResolver) Type(ctx context.Context, obj *claimsstore.ContentCredential) ([]string, error) {
	ts := make([]string, len(obj.Type))
	for ind, val := range obj.Type {
		ts[ind] = string(val)
	}
	return ts, nil
}
func (r *claimResolver) IssuanceDate(ctx context.Context, obj *claimsstore.ContentCredential) (string, error) {
	return obj.IssuanceDate.Format(timeFormat), nil
}
