#!/bin/bash

# OpenJobs Build Script
# Automatically sets version and build time

# Get version from git tag or use date-based version
if git describe --tags --exact-match 2>/dev/null; then
    VERSION=$(git describe --tags --exact-match)
else
    # Use date-based version: YYYY.MM.DD-HHMM
    VERSION=$(date +"%Y.%m.%d-%H%M")
fi

# Get build time
BUILD_TIME=$(date -u +"%Y-%m-%d %H:%M UTC")

# Get git commit hash (short)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "🔨 Building OpenJobs..."
echo "   Version: $VERSION"
echo "   Build Time: $BUILD_TIME"
echo "   Commit: $GIT_COMMIT"
echo ""

# Build with version info embedded
go build \
    -ldflags "-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME'" \
    -o openjobs \
    cmd/openjobs/main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    echo "Run with: ./openjobs"
    echo "Version: $VERSION"
else
    echo "❌ Build failed!"
    exit 1
fi
