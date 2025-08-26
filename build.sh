#!/bin/bash

# This script builds the stt-app executable.
# Use the --app flag to build a bundled macOS .app application.

# Exit immediately if a command exits with a non-zero status.
set -e

BUILD_DIR="builds"

# Create the builds directory if it doesn't exist
mkdir -p "$BUILD_DIR"

if [ "$1" == "--app" ]; then
    ### Build the .app bundle ###
    APP_NAME="stt-app.app"
    APP_PATH="$BUILD_DIR/$APP_NAME"

    echo "==> Building bundled macOS application..."

    echo "--> Cleaning up previous build..."
    rm -rf "$APP_PATH"

    echo "--> Creating .app directory structure..."
    mkdir -p "$APP_PATH/Contents/MacOS"

    echo "--> Copying Info.plist..."
    cp Info.plist "$APP_PATH/Contents/"

    echo "--> Building stt-app executable..."
    go build -ldflags="-w -s" -o "$APP_PATH/Contents/MacOS/stt-app" .

    echo "==> Build complete! Application is at $APP_PATH"
else
    ### Build only the binary ###
    echo "==> Building binary..."
    go build -ldflags="-w -s" -o "$BUILD_DIR/stt-app" .
    echo "==> Build complete! Executable is at $BUILD_DIR/stt-app"
fi
