FROM golang:1.22 as builder
RUN apt update && \
    apt install -y curl net-tools && \
    curl -sSf https://atlasgo.sh | sh && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && \
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest && \
    go install github.com/lerenn/asyncapi-codegen/cmd/asyncapi-codegen@latest

WORKDIR /app
