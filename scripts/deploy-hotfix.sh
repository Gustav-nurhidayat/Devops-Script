#!/usr/bin/env bash

set -euo pipefail

CONTAINER="${CONTAINER:-simplejourney-hotfix}"
VERSION="${VERSION:-dev}"
PORT="${PORT:-8081}"

echo "==> Deploying version: ${VERSION}"
echo "==> Target container: ${CONTAINER}"

if ! docker inspect "${CONTAINER}" >/dev/null 2>&1; then
    echo "ERROR: container '${CONTAINER}' does not exist."
    exit 1
fi

if [ "$(docker inspect -f '{{.State.Running}}' "${CONTAINER}")" != "true" ]; then
    echo "ERROR: container '${CONTAINER}' is not running."
    exit 1
fi

echo "==> Building hotfix binary..."

mkdir -p deploy

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o deploy/app \
    .

echo "==> Copying binary into running container..."

docker cp deploy/app "${CONTAINER}:/app.new"

echo "==> Triggering hot reload..."

docker kill --signal=SIGHUP "${CONTAINER}" >/dev/null

echo "==> Waiting for application..."

for i in {1..10}; do
    if curl -fsS "http://localhost:${PORT}/" | grep -F "version=${VERSION}" >/dev/null; then
        echo "==> Deployment successful."
        curl -fsS "http://localhost:${PORT}/"
        exit 0
    fi

    sleep 1
done

echo "ERROR: deployment health check failed."
exit 1