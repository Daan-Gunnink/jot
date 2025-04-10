#!/bin/bash

# Get version from argument or use default
VERSION=${1:-"0.1.0"}
echo "Building with version $VERSION"

# Build with ldflags to inject version
wails build -ldflags "-X 'main.Version=$VERSION'"

echo "Build complete with version $VERSION" 