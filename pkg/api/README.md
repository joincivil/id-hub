# ID Hub API

The API for ID Hub is gRPC with generated REST and GraphQL proxies.  All of the API is generated off the `.proto` files, which are broken up into logical sections of the API.

## Development

### Install Tooling

```
# Install protoc and protoc plugins for code generation
make setup-api-tools
```

### .proto Files

The `.proto` files defines the `services` and `messages` for the gRPC API. We are using `proto3`.

Edit the `pkg/api/*.proto` files and re-generate the code using the commands below. The generation will generate a set of files for each proto files.  The services generated can be added to the gRPC / HTTP via `Register*Server` calls.

Some good places to learn more about protobufs:

* [https://developers.google.com/protocol-buffers/docs/proto3](https://developers.google.com/protocol-buffers/docs/proto3)
* [https://developers.google.com/protocol-buffers/docs/reference/go-generated](https://developers.google.com/protocol-buffers/docs/reference/go-generated)


### Generate API code


```
# Generate go code from .proto
make generate-protobufs

# Generate HTTP proxy code from .proto
make generate-http-proxy

# Generate GraphQL proxy code form .proto
make generate-gql-proxy

# Does some magic on gqlgen yml files and runs gqlgen on generated yml
make generate-gql

OR

# run all of these with one command
make generate-api

```

Generally, unless you are adding a new `.proto` file, you can just run `make generate-api` on updates to existing `.proto` files.

The generated code will go into `pkg/generated/api`.

### If you add new .proto files

Once the code from the new `.proto` is generated from the above section, it must be added to `pkg/api/resolver.go` and registered on server startup.

In `pkg/idhubmain/server.go`, register the new HTTP and gRPC handlers via the `Register*HandlerServer` and `Register*Server` in `runHTTPProxy` and `RunServer`

```
err := gapi.RegisterDidServiceHandlerServer(ctx, mux, didServer)
...
err = gapi.RegisterClaimServiceHandlerServer(ctx, mux, claimServer)
...
// <- add here

...

gapi.RegisterDidServiceServer(grpcServer, didServer)
gapi.RegisterClaimServiceServer(grpcServer, claimServer)
// <- add here
```

For GraphQL support, embed the newly generated `*GQLServer` struct into the `Resolver` type in `pkg/api/resolver.go`.

```
type Resolver struct {
	*gapi.DidServiceGQLServer
	*gapi.ClaimServiceGQLServer
	// <- add here
}
```
Then ensure we add the `*GQLServer` to the `Resolver` on initialization in `server.go`

```
didGqlServer := &gapi.DidServiceGQLServer{
	Service: didServer,
}
claimGqlServer := &gapi.ClaimServiceGQLServer{
	Service: claimServer,
}
...
resolver := &api.Resolver{
	DidServiceGQLServer:   didGqlServer,
	ClaimServiceGQLServer: claimGqlServer,
	// <- add here
}
```

If you are adding a new .proto file, also read up on the section below to add handling for GraphQL.

### Notes on generate-gql

Running `make generate-gql` will perform some magic to merge all the gqlgen `.yml` files derived from the `.proto` files into a central `gqlgen.yml` to configure `gqlgen` to build a single `exec.go`.

This command will derive the central `gqlgen.yml` file from `pkg/api/gqlgen.tmpl.yml`, so any common attributes should be added to this file before `generate-gql` is run. Same for `pkg/api/gqlgen.tmpl.graphqls`, which is the template for `gqlgen.graphqls`.

All the `*.gqlgen.pb.yml` files will be merged into `gqlgen.yml` file using the merge function of the `yq` command.

This will also alter `.graphqls` files to add an `extend` to the `Query` and `Mutation` types so `gqlgen` will merge all the queries and mutations together.

At this point, `make generate-gql` can be run to run gqlgen again with the newly merged yml config.
