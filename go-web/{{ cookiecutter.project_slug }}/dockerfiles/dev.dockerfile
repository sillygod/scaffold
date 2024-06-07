FROM golang:1.22 as builder
RUN apt update && apt install -y curl net-tools && go install github.com/go-delve/delve/cmd/dlv@latest
WORKDIR /app
