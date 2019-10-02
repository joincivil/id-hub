
![Civil Logo](docs/civil_logo_white.png?raw=true)

---
# Civil ID Hub

[![CircleCI](https://img.shields.io/circleci/project/github/joincivil/id-hub.svg)](https://circleci.com/gh/joincivil/id-hub/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/joincivil/id-hub)](https://goreportcard.com/report/github.com/joincivil/id-hub)


## Contributing

The Civil ID Hub is free and open-source. We are looking to evolve this into something the identity ecosystem will find helpful and effortless to use. We encourage your input via PRs, issues and general communication. Please don't be shy.

## Install Requirements

This project is using `make` to run setup, builds, tests, etc and has been tested and running on `go 1.12.7`.  This repo supports go modules so adding it to your `GOPATH` is unnecessary.

To setup the necessary requirements:

```
make setup
```

## Lint

Check all the packages for linting errors using a variety of linters via `golangci-lint`.  Check the `Makefile` for the up to date list of linters.

```
make lint
```

## Build


```
make build
```

## Testing

Runs the tests and checks code coverage across the project. Produces a `coverage.txt` file for use later.

```
make test
```

## Code Coverage Tool

Run `make test` and launches the HTML code coverage tool.

```
make cover
```

## Run

The Civil ID Hub relies on environment vars for configuration. To configure locally, edit the `.env` file included in the repo to what is needed.

To run the service:

```
go run cmd/idhub/main.go
```

To find all the available configuration environment vars:

```
go run cmd/idhub/main.go -h
```

There is a CLI script that runs various commands. Run this script to view the available commands:

```
go run cmd/idhubcli/main.go -h
```

### Supported Persister Types
`none`, `postgresql`

### Enable Info Logging

Add `-logtostderr=true -stderrthreshold=INFO -v=2` as arguments for the `main.go` command.


