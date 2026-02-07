#!/usr/bin/env bash
set -eo pipefail

echo "Building for all platforms..."

# Handle GITHUB_STEP_SUMMARY if running locally
if [ -z "$GITHUB_STEP_SUMMARY" ]; then
    export GITHUB_STEP_SUMMARY="/dev/null"
fi

mkdir -p build

echo "# Built targets" >> "$GITHUB_STEP_SUMMARY"

export CGO_ENABLED=0
go tool dist list | grep -v wasm | while IFS=/ read -r GOOS GOARCH; do
  echo "Building $GOOS/$GOARCH..."

  if [ "$GITHUB_ACTIONS" == "true" ]; then
      echo "::group::Build $GOOS/$GOARCH"
  fi

  GOOS=$GOOS GOARCH=$GOARCH go build -v -o build/redial_proxy-$GOOS-$GOARCH ./cmd/redial_proxy && (echo "- $GOOS/$GOARCH" >> "$GITHUB_STEP_SUMMARY") || true

  if [ "$GITHUB_ACTIONS" == "true" ]; then
      echo "::endgroup::"
  fi
done
