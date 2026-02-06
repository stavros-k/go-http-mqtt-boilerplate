#!/bin/bash
set -e
OUTPUT_BINARY=./data/server

echo "Running golangci-lint..."
golangci-lint run --fix || true

echo "Running go fix..."
go fix ./...

echo "Running sqlc commands..."
sqlc vet
sqlc compile
sqlc generate

echo "Starting server for code generation..."
go build -o ${OUTPUT_BINARY} ./backend/cmd/server
LOG_LEVEL=debug GENERATE=true ${OUTPUT_BINARY}

cp api_local/api_docs.json .
npm run fmt

echo "Running docs build..."
npm run docs:build

echo "Building server binary..."
go build -o ${OUTPUT_BINARY} ./backend/cmd/server
