package idhubmain

import (
	"context"
	"net"
	"net/http"

	log "github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/joincivil/id-hub/pkg/api"

	grpc "google.golang.org/grpc"
	grpcreflect "google.golang.org/grpc/reflection"
)

func runHTTPProxy() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := api.RegisterIdHubHandlerFromEndpoint(ctx, mux, "localhost:9000", opts)
	if err != nil {
		return err
	}

	log.Infof("Starting up HTTP proxy at localhost:8080")
	return http.ListenAndServe(":8080", mux)
}

// RunServer runs the ID Hub service
func RunServer() error {
	lis, err := net.Listen("tcp", "localhost:9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	service := &api.IDHubService{}

	grpcServer := grpc.NewServer(opts...)
	api.RegisterIdHubServer(grpcServer, service)
	grpcreflect.Register(grpcServer)

	go func() {
		err := runHTTPProxy()
		if err != nil {
			log.Errorf("Error starting up HTTP proxy: err: %v", err)
		}
	}()

	log.Infof("Starting up ID Hub at localhost:9000")
	return grpcServer.Serve(lis)
}
