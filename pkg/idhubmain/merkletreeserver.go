package idhubmain

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/golang/glog"
	"github.com/joincivil/id-hub/pkg/merkletree"
)

// RunMerkleTreeServer starts the merkle tree service server
func RunMerkleTreeServer() error {
	config := populateConfig()

	// init GORM
	db, err := initGorm(config)
	if err != nil {
		log.Fatalf("error initializing gorm")
	}
	db.LogMode(true)

	router := basicHTTPSetup()

	_, _, claimsService, didJWTService := initServices(db, config)

	mtservice := merkletree.NewService(didJWTService, claimsService)

	handler := merkletree.NewHandler(mtservice)

	router.Route(fmt.Sprintf("/%v/merkletree", "v1"), func(r chi.Router) {
		r.Get("/proof/{credential}", handler.GetProofHandler)
		r.Post("/", handler.AddHandler)
		r.Put("/revoke", handler.RevokeHandler)
	})

	gqlURL := fmt.Sprintf(":%v", config.GqlPort)

	return http.ListenAndServe(gqlURL, router)
}
