package idhubmain

import (
	"context"
	"fmt"
	"net"
	"net/http"

	log "github.com/golang/glog"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/joincivil/id-hub/pkg/api"
	gapi "github.com/joincivil/id-hub/pkg/generated/api"
	"github.com/joincivil/id-hub/pkg/utils"

	grpc "google.golang.org/grpc"
	grpcreflect "google.golang.org/grpc/reflection"
)

// TODO(PN): Needs some serious cleanup, placeholder code.

func runHTTPProxy(config *utils.IDHubConfig, didServer gapi.DidServiceServer,
	claimServer gapi.ClaimServiceServer) error {
	if config.HTTPPort == 0 {
		log.Errorf("HTTP proxy is disabled, port was set to 0")
		return nil
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	err := gapi.RegisterDidServiceHandlerServer(ctx, mux, didServer)
	if err != nil {
		log.Errorf("Error starting up HTTP proxy: err: %v", err)
		return err
	}
	err = gapi.RegisterClaimServiceHandlerServer(ctx, mux, claimServer)
	if err != nil {
		log.Errorf("Error starting up HTTP proxy: err: %v", err)
		return err
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Handle("/*", mux)

	httpURL := fmt.Sprintf("localhost:%v", config.HTTPPort)
	log.Infof("Starting up HTTP proxy at %v", httpURL)
	err = http.ListenAndServe(httpURL, router)
	if err != nil {
		log.Errorf("Error starting GraphQL proxy: %v", err)
		return err
	}

	return nil
}

func runGraphQLProxy(config *utils.IDHubConfig, didServer gapi.DidServiceServer,
	claimServer gapi.ClaimServiceServer) error {
	if config.GqlPort == 0 {
		log.Errorf("GraphQL proxy is disabled, port was set to 0")
		return nil
	}

	didGqlServer := &gapi.DidServiceGQLServer{
		Service: didServer,
	}
	claimGqlServer := &gapi.ClaimServiceGQLServer{
		Service: claimServer,
	}
	resolver := &api.Resolver{
		DidServiceGQLServer:   didGqlServer,
		ClaimServiceGQLServer: claimGqlServer,
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	queryHandler := handler.GraphQL(
		gapi.NewExecutableSchema(
			gapi.Config{Resolvers: resolver},
		),
	)
	router.Handle(
		fmt.Sprintf("/%v/query", "v1"),
		queryHandler,
	)

	gqlURL := fmt.Sprintf("localhost:%v", config.GqlPort)
	log.Infof("Starting up GraphQL proxy at %v", gqlURL)
	err := http.ListenAndServe(gqlURL, router)
	if err != nil {
		log.Errorf("Error starting GraphQL proxy: %v", err)
		return err
	}

	return nil
}

// RunServer runs the ID Hub service
func RunServer() error {
	config := populateConfig()

	db, err := initGorm(config)
	if err != nil {
		log.Fatalf("error initializing gorm")
	}

	didPersister := initDidPersister(db)
	didService := initDidService(didPersister)
	didServer := api.NewDidImplementedServer(didService)

	claimServer := api.NewClaimServiceImplementedServer()

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	gapi.RegisterDidServiceServer(grpcServer, didServer)
	gapi.RegisterClaimServiceServer(grpcServer, claimServer)
	grpcreflect.Register(grpcServer)

	go func() {
		err := runHTTPProxy(config, didServer, claimServer)
		if err != nil {
			log.Errorf("error starting up HTTP proxy")
		}
	}()

	go func() {
		err := runGraphQLProxy(config, didServer, claimServer)
		if err != nil {
			log.Errorf("error starting up GraphQL proxy")
		}
	}()

	grpcURL := fmt.Sprintf("localhost:%v", config.GrpcPort)
	log.Infof("Starting up gRPC at %v", grpcURL)
	lis, err := net.Listen("tcp", grpcURL)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	return grpcServer.Serve(lis)
}
