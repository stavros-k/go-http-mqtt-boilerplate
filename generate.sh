#!/bin/bash
set -e
OUTPUT_BINARY=./data/server

echo "Running go fix..."
go fix ./...

echo "Running sqlc commands..."
sqlc vet
sqlc compile
sqlc generate

echo "Starting server for code generation..."
go build -o ${OUTPUT_BINARY} ./backend/cmd/server
LOG_LEVEL=debug GENERATE=true ${OUTPUT_BINARY}

npm run fmt

echo "Running docs build..."
DOCS_BASE_PATH=/docs npm run docs:build

echo "Building server binary..."
go build -o ${OUTPUT_BINARY} ./backend/cmd/server
