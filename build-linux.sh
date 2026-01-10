#!/bin/bash

# Script to build the project for Linux
# Usage: ./build-linux.sh [docker]
# If 'docker' is passed as argument, uses Docker to compile (recommended)

set -e

# Create build directory if it doesn't exist
BUILD_DIR="linux-build"
mkdir -p "$BUILD_DIR"
OUTPUT_FILE="$BUILD_DIR/whatsmiau-linux"

if [ "$1" = "docker" ]; then
    echo "Building using Docker (recommended)..."
    echo "   Creating static self-contained binary (no system dependencies required)"
    
    # Read WEBHOOK_URL from environment or use empty string
    BUILD_WEBHOOK_URL="${WEBHOOK_URL:-}"
    
    # Create a temporary container to compile with musl (easier to make static)
    docker run --rm -v "$(pwd)":/app -w /app \
        -e CGO_ENABLED=1 \
        -e BUILD_WEBHOOK_URL="$BUILD_WEBHOOK_URL" \
        golang:1.25-alpine \
        sh -c "
            apk add --no-cache gcc musl-dev sqlite-dev build-base && \
            go mod download && \
            mkdir -p $BUILD_DIR && \
            (CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s -X github.com/verbeux-ai/whatsmiau/env.BuildWebhookURL=\"$BUILD_WEBHOOK_URL\" -linkmode external -extldflags \"-static\"' -o $OUTPUT_FILE main.go || \
            CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s -X github.com/verbeux-ai/whatsmiau/env.BuildWebhookURL=\"$BUILD_WEBHOOK_URL\"' -o $OUTPUT_FILE main.go)
        "
    
    if [ $? -eq 0 ]; then
        echo "Build successful with Docker!"
        echo "Executable created: $OUTPUT_FILE"
        ls -lh "$OUTPUT_FILE"
        echo ""
        echo "The executable is ready to use on Linux"
        echo "   Static self-contained binary - NO need to install anything on the server"
        echo "   To transfer: scp $OUTPUT_FILE user@server:/destination/path/"
        echo ""
        echo "On the Linux server you only need:"
        echo "   1. chmod +x whatsmiau-linux"
        echo "   2. ./whatsmiau-linux"
    else
        echo "Error building with Docker"
        exit 1
    fi
else
    echo "Building for Linux (amd64) locally..."
    echo "Note: SQLite requires CGO. If it fails, use: ./build-linux.sh docker"
    echo ""
    
    # Check if CGO is available
    if ! command -v gcc &> /dev/null; then
        echo "Error: gcc is not installed. SQLite requires CGO."
        echo "   Options:"
        echo "   1. Install gcc: brew install gcc (macOS) or apt-get install gcc (Linux)"
        echo "   2. Use Docker: ./build-linux.sh docker"
        exit 1
    fi
    
    # Read WEBHOOK_URL from environment or use empty string
    BUILD_WEBHOOK_URL="${WEBHOOK_URL:-}"
    if [ -n "$BUILD_WEBHOOK_URL" ]; then
        echo "   Injecting WEBHOOK_URL at build time: $BUILD_WEBHOOK_URL"
    fi
    
    # Build with CGO enabled (may require system libraries)
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-w -s -X github.com/verbeux-ai/whatsmiau/env.BuildWebhookURL=$BUILD_WEBHOOK_URL" -o "$OUTPUT_FILE" main.go
    
    if [ $? -eq 0 ]; then
        echo "Build successful!"
        echo "Executable created: $OUTPUT_FILE"
        ls -lh "$OUTPUT_FILE"
        echo ""
        echo "The executable is ready to use on Linux"
        echo "   To transfer: scp $OUTPUT_FILE user@server:/destination/path/"
        echo ""
        echo "Note: This executable may require C libraries on the Linux server."
        echo "   For a completely self-contained binary, use: ./build-linux.sh docker"
    else
        echo "Build error"
        echo ""
        echo "Suggestion: Use Docker to build (more reliable):"
        echo "   ./build-linux.sh docker"
        exit 1
    fi
fi

