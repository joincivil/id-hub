package idhubmain

import (
	"github.com/joincivil/id-hub/pkg/did"
)

func initDidService(persister did.Persister) *did.Service {
	return did.NewService(persister)
}
