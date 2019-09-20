package graphql

import (
	"context"

	"github.com/joincivil/id-hub/pkg/did"
)

const (
	resolverVersion = "1.0"
)

// Resolver is the main GraphQL resolver
type Resolver struct {
	DidService *did.Service
}

// Version returns the version of the GraphQL API
func (r *Resolver) Version(ctx context.Context) (string, error) {
	return resolverVersion, nil
}

// Query is the resolver for the Query type
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Mutation is the resolver for the Mutation type
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

type queryResolver struct{ *Resolver }

type mutationResolver struct{ *Resolver }
