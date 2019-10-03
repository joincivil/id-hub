package graphql

import (
	"context"

	didlib "github.com/ockam-network/did"
	"github.com/pkg/errors"

	"github.com/joincivil/go-common/pkg/article"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/utils"
)

// ResolverRoot

// ArticleMetadata returns the resolver for article metadata
func (r *Resolver) ArticleMetadata() ArticleMetadataResolver {
	return &articleMetadataResolver{r}
}

// ArticleMetadataImage returns resolver for metadata images
func (r *Resolver) ArticleMetadataImage() ArticleMetadataImageResolver {
	return &articleMetadataImageResolver{r}
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

// Claim Resolvers

type articleMetadataResolver struct{ *Resolver }

func (r *articleMetadataResolver) RevisionDate(ctx context.Context, obj *article.Metadata) (*string, error) {
	opd := obj.RevisionDate.Format("2006-01-02T15:04:05Z")
	return utils.StrToPtr(opd), nil
}
func (r *articleMetadataResolver) OriginalPublishDate(ctx context.Context, obj *article.Metadata) (*string, error) {
	opd := obj.OriginalPublishDate.Format("2006-01-02T15:04:05Z")
	return utils.StrToPtr(opd), nil
}

type articleMetadataImageResolver struct{ *Resolver }

func (r *articleMetadataImageResolver) Height(ctx context.Context, obj *article.Image) (*int, error) {
	return utils.IntToPtr(obj.H), nil
}
func (r *articleMetadataImageResolver) Width(ctx context.Context, obj *article.Image) (*int, error) {
	return utils.IntToPtr(obj.W), nil
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
	return obj.IssuanceDate.Format("2006-01-02T15:04:05Z"), nil
}
