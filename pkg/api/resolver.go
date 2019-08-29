package api

import (
	"context"

	gapi "github.com/joincivil/id-hub/pkg/generated/api"
)

const (
	resolverVersion = "1.0"
)

// Resolver is the main GraphQL resolver
type Resolver struct {
	*gapi.DidServiceGQLServer
	*gapi.ClaimServiceGQLServer
}

// Version returns the version of the GraphQL API
func (r *Resolver) Version(ctx context.Context) (string, error) {
	return resolverVersion, nil
}

// Query is the resolver for the Query type
func (r *Resolver) Query() gapi.QueryResolver {
	return &queryResolver{r}
}

// Mutation is the resolver for the Mutation type
func (r *Resolver) Mutation() gapi.MutationResolver {
	return &mutationResolver{r}
}

type queryResolver struct{ *Resolver }

type mutationResolver struct{ *Resolver }
