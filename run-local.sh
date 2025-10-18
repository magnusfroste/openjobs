#!/bin/bash

# Run OpenJobs locally with all fixes
# This connects to your production Supabase but runs the dashboard locally

echo "🚀 Starting OpenJobs locally..."
echo ""
echo "📊 Dashboard will be available at: http://localhost:9090/dashboard"
echo "🔌 API will be available at: http://localhost:9090"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Set environment variables
export SUPABASE_URL=https://supabase.froste.eu
export SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyAgCiAgICAicm9sZSI6ICJhbm9uIiwKICAgICJpc3MiOiAic3VwYWJhc2UtZGVtbyIsCiAgICAiaWF0IjogMTY0MTc2OTIwMCwKICAgICJleHAiOiAxNzk5NTM1NjAwCn0.dc_X5iR_VP_qT0zsiyj_I_OZ2T9FtRU2BBNWN8Bu4GE
export PORT=9090

# Microservices mode - point to production plugin containers
export USE_HTTP_PLUGINS=true
export PLUGIN_ARBETSFORMEDLINGEN_URL=https://app-openjobs-arbfrm.katsu6.easypanel.host
export PLUGIN_EURES_URL=https://app-openjobs-eures.katsu6.easypanel.host
export PLUGIN_REMOTIVE_URL=https://app-openjobs-remotive.katsu6.easypanel.host
export PLUGIN_REMOTEOK_URL=https://app-openjobs-remoteok.katsu6.easypanel.host

# Build with version
VERSION=$(date +"%Y.%m.%d-%H%M")
BUILD_TIME=$(date -u +"%Y-%m-%d %H:%M UTC")

echo "Building OpenJobs v$VERSION..."
go build \
    -ldflags "-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME'" \
    -o openjobs \
    cmd/openjobs/main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    ./openjobs
else
    echo "❌ Build failed!"
    exit 1
fi
