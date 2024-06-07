FROM golang:1.22 as builder
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main main.go


# use a minimal alpine image
FROM alpine:3.20
COPY --from=builder /app/main /app/main
ENV APP_ENV=dev
ENV PORT=8080

WORKDIR /app

HEALTHCHECK --interval=5s --timeout=10s --start-period=5s \
  CMD curl -fs http://localhost:$PORT/health || exit 1

CMD ["./main"]
