# {{ cookiecutter.project_name }}

## dev dependencies

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
atlas schema apply --url "sqlite://file.db" --env local
```

{% raw %}
```sh
atlas migrate diff initial --env local
```
{% endraw %}

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
