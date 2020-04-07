package idhubmain

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/joincivil/id-hub/pkg/hedgehog"

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
	"github.com/rs/cors"

	"github.com/vektah/gqlparser/gqlerror"
)

var (
	validCorsOrigins = []string{
		"*",
	}
)

func initResolver(db *gorm.DB, config *utils.IDHubConfig) *graphql.Resolver {

	jwtService, didService, claimsService, _ := initServices(db, config)

	return &graphql.Resolver{
		DidService:   didService,
		ClaimService: claimsService,
		JWTService:   jwtService,
	}
}

func basicHTTPSetup() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	cors := cors.New(cors.Options{
		AllowedOrigins:   validCorsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
		Debug:            true,
	})
	router.Use(cors.Handler)
	return router
}

// RunServer runs the ID Hub service
func RunServer() error {
	config := populateConfig()

	// init GORM
	db, err := initGorm(config)
	if err != nil {
		log.Fatalf("error initializing gorm")
	}
	// db.LogMode(true)

	resolver := initResolver(db, config)
	initHedgehog(db)

	router := basicHTTPSetup()

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

	router.Handle("/", handler.Playground("GraphQL playground",
		fmt.Sprintf("/%v/%v", "v1", "query")))
	gqlURL := fmt.Sprintf(":%v", config.GqlPort)

	hedgehog.AddRoutes(hedgehog.Dependencies{Router: router, Db: db})

	log.Infof("Starting up GraphQL services at %v", gqlURL)
	return http.ListenAndServe(gqlURL, router)
}
