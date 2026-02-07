#!/bin/bash
set -e
mkdir -p build
echo "# Built targets"
export CGO_ENABLED=0
go tool dist list | grep -v wasm | while IFS=/ read -r GOOS GOARCH; do
  echo "Building $GOOS/$GOARCH"
  GOOS=$GOOS GOARCH=$GOARCH go build -v -o build/redial_proxy-$GOOS-$GOARCH ./cmd/redial_proxy
done
