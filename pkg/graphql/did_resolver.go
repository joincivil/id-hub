package graphql

import (
	"context"

	"github.com/pkg/errors"

	"github.com/joincivil/id-hub/pkg/utils"

	"github.com/joincivil/id-hub/pkg/did"
)

// ResolverRoot

// DidDocAuthentication is the resolver for DID Authentications
func (r *Resolver) DidDocAuthentication() DidDocAuthenticationResolver {
	return &didDocAuthenticationResolver{r}
}

// DidDocPublicKey is the resolver for the DID public key
func (r *Resolver) DidDocPublicKey() DidDocPublicKeyResolver {
	return &didDocPublicKeyResolver{r}
}

// DidDocService is the resolver for the DID service
func (r *Resolver) DidDocService() DidDocServiceResolver {
	return &didDocServiceResolver{r}
}

// DidDocument is the resolver for the DID document
func (r *Resolver) DidDocument() DidDocumentResolver {
	return &didDocumentResolver{r}
}

// Queries

func (r *queryResolver) DidGet(ctx context.Context, in *DidGetRequestInput) (
	*DidGetResponse, error) {
	if in.Did == nil {
		return nil, errors.New("did is empty")
	}

	doc, err := r.DidService.GetDocument(*in.Did)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve did document")
	}

	return &DidGetResponse{Doc: doc}, nil
}

// Did Resolvers

type didDocAuthenticationResolver struct{ *Resolver }

func (r *didDocAuthenticationResolver) PublicKey(ctx context.Context, obj *did.DocAuthenicationWrapper) (
	*did.DocPublicKey, error) {
	return &obj.DocPublicKey, nil
}

type didDocPublicKeyResolver struct{ *Resolver }

func (r *didDocPublicKeyResolver) ID(ctx context.Context, obj *did.DocPublicKey) (*string, error) {
	if obj.ID != nil {
		val := obj.ID.String()
		return &val, nil
	}
	return nil, nil
}
func (r *didDocPublicKeyResolver) Type(ctx context.Context, obj *did.DocPublicKey) (*string, error) {
	val := string(obj.Type)
	return &val, nil
}
func (r *didDocPublicKeyResolver) Controller(ctx context.Context, obj *did.DocPublicKey) (*string, error) {
	if obj.Controller != nil {
		val := obj.Controller.String()
		return &val, nil
	}
	return nil, nil
}

type didDocServiceResolver struct{ *Resolver }

func (r *didDocServiceResolver) ID(ctx context.Context, obj *did.DocService) (*string, error) {
	val := obj.ID.String()
	return &val, nil
}
func (r *didDocServiceResolver) ServiceEndpoint(ctx context.Context, obj *did.DocService) (*utils.AnyValue, error) {
	return &utils.AnyValue{
		Value: obj.ServiceEndpoint,
	}, nil
}

type didDocumentResolver struct{ *Resolver }

func (r *didDocumentResolver) ID(ctx context.Context, obj *did.Document) (*string, error) {
	val := obj.ID.String()
	return &val, nil
}
func (r *didDocumentResolver) Controller(ctx context.Context, obj *did.Document) (*string, error) {
	if obj.Controller != nil {
		val := obj.Controller.String()
		return &val, nil
	}
	return nil, nil
}
