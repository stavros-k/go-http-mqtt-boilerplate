#!/bin/bash
set -e

mkdir -p data
component="$1"
if [ "$component" == "local" ]; then
  echo "Component is local"
  OUTPUT_BINARY=./data/local
  JSON_DOCS_FILE=./api_local/api_docs.json

elif [ "$component" == "cloud" ]; then
  echo "Component is cloud"
  OUTPUT_BINARY=./data/cloud
  JSON_DOCS_FILE=./api_cloud/api_docs.json

else
  echo "Usage: $0 [local|cloud]"
  exit 1
fi

rm api_docs.json || echo "No existing api_docs.json to remove."

echo "Running golangci-lint..."
golangci-lint run --fix || true

echo "Running go fix..."
go fix ./...


echo "Building [${component}] server for code generation..."
go build -o ${OUTPUT_BINARY} ./backend/cmd/${component}
echo "Running [${component}] server for code generation..."
LOG_LEVEL=debug GENERATE=true ${OUTPUT_BINARY}

echo "Running sqlc commands..."
sqlc vet
sqlc compile
sqlc generate

echo "Copying [${JSON_DOCS_FILE}] to [api_docs.json]..."
cp ${JSON_DOCS_FILE} api_docs.json

echo "Running formatter..."
npm run fmt

echo "Running docs build..."
npm run docs:build

echo "Building server binary with the docs..."
go build -o ${OUTPUT_BINARY} ./backend/cmd/${component}
