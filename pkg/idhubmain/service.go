package idhubmain

import (
	"github.com/iden3/go-iden3-core/db"
	"github.com/joincivil/id-hub/pkg/claims"
	"github.com/joincivil/id-hub/pkg/claimsstore"
	"github.com/joincivil/id-hub/pkg/did"
)

func initDidService(persister did.Persister) *did.Service {
	return did.NewService(persister)
}

func initClaimsService(treeStore db.Storage, signedClaimStore *claimsstore.SignedClaimPGPersister,
	didService *did.Service) (*claims.Service, error) {
	return claims.NewService(treeStore, signedClaimStore, didService)
}
