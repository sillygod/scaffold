## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


.PHONY: install-dependencies
	go env GOOS
	# todo add installation of dependencies


## run: run the application
.PHONY: run
run: sqlc oapi async-api build
	./main

.PHONY: run-worker
run-worker:
	go build -o main-worker cmd/worker/main.go
	./main-worker

.PHONY: run-scheduler
run-scheduler:
	go build -o main-scheduler cmd/scheduler/main.go
	./main-scheduler

.PHONY: run-ws
run-ws:
	go build -o main-ws cmd/ws/main.go
	./main-ws

## sqlc: generate sqlc queries
.PHONY: sqlc
sqlc:
	sqlc generate

## oapi: generate oapi schema
.PHONY: oapi
oapi:
	oapi-codegen -config cfg.yaml swagger.yml

.PHONY: async-api
async-api:
	asyncapi-codegen -i asyncapi.yaml -p events -o events/asyncapi.gen.go

## build: build the application
.PHONY: build
build:
	go build -o main

## tidy: run go fmt and go mod tidy
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v


## audit: run go mod verify, go vet and some static checks
.PHONY: audit
audit:
	go mod verify
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go vet ./...

## test: run tests
.PHONY: test
test:
	go test -v -race ./...
