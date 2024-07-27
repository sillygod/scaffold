# {{ cookiecutter.project_name }}

## dev dependencies

```sh
go mod download
```

```sh
cp .env.example .env
```



a nice database migration tool
https://github.com/ariga/atlas

```sh
brew install ariga/tap/atlas
```

atlas schema inspect -u "sqlite://file?cache=shared&mode=memory"

https://atlasgo.io/declarative/inspect#examples

### apply migrations

example usages

```sh
atlas schema apply --env docker-local
```

```sh
atlas schema clean --env docker-local
```

```sh
atlas migrate diff initial --env docker-local
```


a type-safe sql generator
https://github.com/sqlc-dev/sqlc

```sh
brew install sqlc
```

## generate db schema

```sh
sqlc generate
```

## how to write the sqlc query annotations
https://docs.sqlc.dev/en/latest/reference/query-annotations.html

## core files

they are responsible for resources lifecycle and some core functions.

## generate restful schemas

follow the schema-first design principle.

```sh
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
# Remember to run the command below when updating the swagger.yml
# this will generate the schema go file to the specific path (current is routers/schemas/schemas.go)
oapi-codegen -config cfg.yaml swagger.yml
```

https://openapi.tools/

schema online editor: https://editor-next.swagger.io/ or this one https://www.apibldr.com/source
openapi guide:  https://swagger.io/docs/specification/about

## serve the doc server

```sh
docker run -p 8888:8080 -e SWAGGER_JSON=/app/swagger.yml -v .:/app swaggerapi/swagger-ui
```

### generate event schemas

follow the schema-first design principle.

https://studio.asyncapi.com/

```sh
brew install asyncapi
go install github.com/lerenn/asyncapi-codegen/cmd/asyncapi-codegen@latest
```

asyncapi-codegen -i asyncapi.yaml -p events -o events/asyncapi.gen.go

### spawn the server

```sh
make run
```

or run in docker container with docker-compose

> all the prerequisites are required above can be ignored, if you are running on docker container.

```sh
docker-compose run -it --rm {{ cookiecutter.project_name }} /bin/bash
make run
```


### debug when running tests

```sh
dlv test --wd . ./tests  -- -test.v
```
