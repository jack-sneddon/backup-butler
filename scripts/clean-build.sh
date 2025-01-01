#!/bin/bash

# Get the directory where the script is located
SCRIPT_DIR="${0:a:h}"
# Get project root (parent of scripts directory)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
# Binary location
BINARY="$PROJECT_ROOT/backup-butler"

cd "$PROJECT_ROOT" || exit 1

echo "Cleaning old build artifacts..."
rm -f "$BINARY"

echo "Formatting Go source code..."
go fmt ./...

echo "Tidying up dependencies..."
go mod tidy

echo "Verifying module integrity..."
go mod verify

echo "Cleaning Go module cache..."
go clean -cache -modcache

echo "Building the project..."
# Build to project root directory
go build -o "$BINARY" cmd/main.go
chmod +x "$BINARY"

echo "Build complete. Binary location: $BINARY"