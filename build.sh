#!/bin/bash

# This script builds the stt-app executable and places it in the builds/ directory.

# Exit immediately if a command exits with a non-zero status.
set -e

echo "==> Creating builds directory..."
mkdir -p builds

echo "==> Building stt-app..."
go build -ldflags="-w -s" -o builds/stt-app .

echo "==> Build complete! Executable is at builds/stt-app"
