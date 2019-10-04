package idhubmain

import (
	"fmt"
	"net/http"

	log "github.com/golang/glog"
	"github.com/joincivil/id-hub/pkg/graphql"
	"github.com/joincivil/id-hub/pkg/utils"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func initResolver(config *utils.IDHubConfig) *graphql.Resolver {
	// init GORM
	db, err := initGorm(config)
	if err != nil {
		log.Fatalf("error initializing gorm")
	}

	// DID init
	didPersister := initDidPersister(db)
	didService := initDidService(didPersister)

	// Claims init
	treePersister := initTreePersister(db)
	signedClaimPersister := initSignedClaimPersister(db)
	claimsService, err := initClaimsService(
		treePersister,
		signedClaimPersister,
		didService,
	)
	if err != nil {
		log.Fatalf("error initializing claims service")
	}

	return &graphql.Resolver{
		DidService:   didService,
		ClaimService: claimsService,
	}
}

// RunServer runs the ID Hub service
func RunServer() error {
	config := populateConfig()

	resolver := initResolver(config)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	queryHandler := handler.GraphQL(
		graphql.NewExecutableSchema(
			graphql.Config{Resolvers: resolver},
		),
	)
	router.Handle(
		fmt.Sprintf("/%v/query", "v1"),
		queryHandler,
	)

	gqlURL := fmt.Sprintf(":%v", config.GqlPort)

	log.Infof("Starting up GraphQL services at %v", gqlURL)
	return http.ListenAndServe(gqlURL, router)
}
