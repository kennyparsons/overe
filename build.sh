#!/bin/bash

# This script builds the stt-app executable.

# Exit immediately if a command exits with a non-zero status.
set -e

BUILD_DIR="builds"

# Create the builds directory if it doesn't exist
mkdir -p "$BUILD_DIR"

### Build only the binary ###
echo "==> Building binary..."
go build -ldflags="-X main.version=dev -w -s -extldflags '-sectcreate __TEXT __info_plist Info.plist'" -o "$BUILD_DIR/stt-app" .
echo "==> Build complete! Executable is at $BUILD_DIR/stt-app"
