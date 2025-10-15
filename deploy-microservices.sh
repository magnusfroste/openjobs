#!/bin/bash

# OpenJobs Microservices Deployment Script
# Deploys core API and plugin services separately

set -e

echo "🚀 Deploying OpenJobs Microservices Architecture"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${GREEN}[STEP]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_step "Checking prerequisites..."

    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi

    echo "✅ Prerequisites check passed"
}

# Build and deploy core API
deploy_core_api() {
    print_step "Building and deploying Core API..."

    # Stop any existing containers
    docker-compose down 2>/dev/null || true

    # Build core API
    docker build -t openjobs-core:latest -f Dockerfile .

    # Start core API
    docker run -d \
        --name openjobs-core \
        --env-file .env \
        -e PLUGIN_ARBETSFORMEDLINGEN_URL=http://openjobs-plugin-af:8081 \
        -e PLUGIN_EURES_URL=http://openjobs-plugin-eures:8082 \
        -p 8080:8080 \
        openjobs-core:latest

    echo "✅ Core API deployed on port 8080"
}

# Deploy Arbetsförmedlingen plugin
deploy_af_plugin() {
    print_step "Building and deploying Arbetsförmedlingen Plugin..."

    # Build plugin
    docker build -t openjobs-plugin-arbetsformedlingen:latest -f Dockerfile.plugin-arbetsformedlingen .

    # Start plugin
    docker run -d \
        --name openjobs-plugin-af \
        --env-file .env \
        -e PORT=8081 \
        -p 8081:8081 \
        openjobs-plugin-arbetsformedlingen:latest

    echo "✅ Arbetsförmedlingen plugin deployed on port 8081"
}

# Deploy EURES plugin
deploy_eures_plugin() {
    print_step "Building and deploying EURES Plugin..."

    # Build plugin
    docker build -t openjobs-plugin-eures:latest -f Dockerfile.plugin-eures .

    # Start plugin
    docker run -d \
        --name openjobs-plugin-eures \
        --env-file .env \
        -e PORT=8082 \
        -p 8082:8082 \
        openjobs-plugin-eures:latest

    echo "✅ EURES plugin deployed on port 8082"
}

# Test deployment
test_deployment() {
    print_step "Testing deployment..."

    echo "⏳ Waiting for services to be ready..."
    sleep 10

    # Test core API
    echo "  Testing Core API..."
    if curl -f -s http://localhost:8080/health > /dev/null; then
        echo "  ✅ Core API is healthy"
    else
        print_error "Core API health check failed"
        show_logs
        exit 1
    fi

    # Test Arbetsförmedlingen plugin
    echo "  Testing Arbetsförmedlingen Plugin..."
    if curl -f -s http://localhost:8081/health > /dev/null; then
        echo "  ✅ Arbetsförmedlingen plugin is healthy"
    else
        print_error "Arbetsförmedlingen plugin health check failed"
        show_logs
        exit 1
    fi

    # Test EURES plugin
    echo "  Testing EURES Plugin..."
    if curl -f -s http://localhost:8082/health > /dev/null; then
        echo "  ✅ EURES plugin is healthy"
    else
        print_error "EURES plugin health check failed"
        show_logs
        exit 1
    fi

    # Test job sync
    echo "  Testing manual sync..."
    if curl -f -s -X POST http://localhost:8080/sync/manual > /dev/null; then
        echo "  ✅ Manual sync completed"
    else
        print_warning "Manual sync call failed (may be expected on first run)"
    fi
}

# Show logs for debugging
show_logs() {
    print_step "Showing service logs for debugging..."

    echo "Core API logs:"
    docker logs openjobs-core 2>/dev/null || echo "No logs available"
    echo ""

    echo "Arbetsförmedlingen plugin logs:"
    docker logs openjobs-plugin-af 2>/dev/null || echo "No logs available"
    echo ""

    echo "EURES plugin logs:"
    docker logs openjobs-plugin-eures 2>/dev/null || echo "No logs available"
}

# Show summary
show_summary() {
    print_step "Deployment Summary"

    echo "🌐 Service URLs:"
    echo "  Core API:           http://localhost:8080"
    echo "  Arbetsförmedlingen: http://localhost:8081"
    echo "  EURES Plugin:       http://localhost:8082"
    echo ""

    echo "📊 Useful commands:"
    echo "  View Core API:      curl http://localhost:8080/health"
    echo "  View jobs:          curl http://localhost:8080/jobs"
    echo "  Manual sync:        curl -X POST http://localhost:8080/sync/manual"
    echo "  Plugin AF health:   curl http://localhost:8081/health"
    echo "  Plugin EURES health:curl http://localhost:8082/health"
    echo ""

    echo "🛑 Stop services:"
    echo "  docker-compose down  # or manually stop each container"
    echo ""

    echo "🎉 All services deployed successfully!"
    echo "   Core API + 2 Plugin containers running"
}

# Main deployment function
main() {
    check_prerequisites
    deploy_core_api
    deploy_af_plugin
    deploy_eures_plugin
    test_deployment
    show_summary
}

# Handle command line arguments
if [ "$1" = "logs" ]; then
    show_logs
    exit 0
elif [ "$1" = "stop" ]; then
    print_step "Stopping all services..."
    docker-compose down 2>/dev/null || docker stop openjobs-core openjobs-plugin-af openjobs-plugin-eures 2>/dev/null || echo "No running containers"
    echo "✅ All services stopped"
    exit 0
fi

# Run main deployment
main
