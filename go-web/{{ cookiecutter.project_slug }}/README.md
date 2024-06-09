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
