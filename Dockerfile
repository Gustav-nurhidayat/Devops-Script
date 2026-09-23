FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod ./
COPY main.go ./
COPY main_test.go ./
COPY supervisor ./supervisor

ARG VERSION=dev

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /app \
    .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w" \
    -o /supervisor \
    ./supervisor

FROM scratch

COPY --from=builder /supervisor /supervisor
COPY --from=builder /app /app

EXPOSE 8080

ENTRYPOINT ["/supervisor"]