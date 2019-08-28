package api

import (
	context "context"

	log "github.com/golang/glog"
)

// IDHubService implements the IdHubServer interface
type IDHubService struct {
}

// GetDID returns the DID Document given the DID
func (i *IDHubService) GetDID(ctx context.Context, d *Did) (*DidDocument, error) {
	log.Infof("GetDID: %v\n", d.Did)
	return &DidDocument{
		Id: "did:ethuri:12345",
	}, nil
}
