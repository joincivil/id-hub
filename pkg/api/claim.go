package api

import (
	context "context"

	gapi "github.com/joincivil/id-hub/pkg/generated/api"
)

// NewClaimServiceImplementedServer is a convenience function that returns a
// new ClaimServiceImplementedServer given it's dependencies
func NewClaimServiceImplementedServer() *ClaimServiceImplementedServer {
	return &ClaimServiceImplementedServer{}
}

// ClaimServiceImplementedServer implements the ClaimServiceServer interface
type ClaimServiceImplementedServer struct {
}

// Get retrieves the claim given the claim request
func (c *ClaimServiceImplementedServer) Get(ctx context.Context, req *gapi.ClaimGetRequest) (
	*gapi.ClaimGetResponse, error) {
	return &gapi.ClaimGetResponse{
		Claim: &gapi.Claim{
			Id:      "testID",
			Context: "@context",
			Types:   []string{"test1", "test2"},
		},
	}, nil
}
