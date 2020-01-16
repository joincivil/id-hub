package idhubmain

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/joincivil/id-hub/pkg/auth"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/graphql"
	"github.com/joincivil/id-hub/pkg/utils"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/vektah/gqlparser/gqlerror"
)

func initResolver(config *utils.IDHubConfig) *graphql.Resolver {
	// init GORM
	db, err := initGorm(config)
	if err != nil {
		log.Fatalf("error initializing gorm")
	}
	// db.LogMode(true)

	// DID init
	// Universal Resolver
	resolver, err := initHTTPUniversalResolver(config)
	if err != nil {
		log.Fatalf("error initializing universal resolver")
	}
	// EthURI Resolver
	ethURIResolver, err := initEthURIResolver(db)
	if err != nil {
		log.Fatalf("error initializing ethuri resolver")
	}
	// TODO(PN): Adding ethuri resolver during transition of enterprise clients
	// to other DID methods. Once this occurs, should remove it.
	didService := initDidService([]did.Resolver{resolver, ethURIResolver})

	// Claims init
	treePersister := initTreePersister(db)
	signedClaimPersister := initSignedClaimPersister(db)
	rootPersister := initRootClaimPersister(db)
	ethHelper, err := initETHHelper(config)
	if err != nil {
		log.Fatalf("error initializing eth helper: %v", err)
	}
	rootService, err := initRootService(config, ethHelper, treePersister, rootPersister)
	if err != nil {
		log.Fatalf("error initializing root service: %v", err)
	}
	dlock := initDLock(config)
	claimsService, err := initClaimsService(
		treePersister,
		signedClaimPersister,
		didService,
		rootService,
		dlock,
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

	// Setup the ID Hub auth middleware
	router.Use(auth.Middleware())

	queryHandler := handler.GraphQL(
		graphql.NewExecutableSchema(
			graphql.Config{Resolvers: resolver},
		),
		handler.ErrorPresenter(
			func(ctx context.Context, e error) *gqlerror.Error {
				err := errors.Cause(e)
				if err == did.ErrResolverDIDNotFound {
					log.Errorf("gql error: err: %v, cause: %v", e, err)
				} else {
					log.Errorf("gql error: err: %+v, cause: %+v", e, err)
				}
				return gqlgen.DefaultErrorPresenter(ctx, err)
			},
		),
		handler.RecoverFunc(func(ctx context.Context, err interface{}) error {
			switch val := err.(type) {
			case error:
				log.Errorf("gql panic error: err: %+v, cause: %+v", val, errors.Cause(val))
			}
			return fmt.Errorf("Internal server error: %v", err)
		}),
	)
	router.Handle(
		fmt.Sprintf("/%v/query", "v1"),
		queryHandler,
	)

	gqlURL := fmt.Sprintf(":%v", config.GqlPort)

	log.Infof("Starting up GraphQL services at %v", gqlURL)
	return http.ListenAndServe(gqlURL, router)
}
