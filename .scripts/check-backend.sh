#!/bin/bash

set -e

echo "=== Checking Go Backend ==="
cd "$(dirname "$0")/../backend"

echo "Running go fmt..."
go fmt ./...

echo "Running go vet..."
go vet ./...

echo "Running go build..."
go build -o /tmp/test-build ./cmd/server

echo "✅ All checks passed!"
