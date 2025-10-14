#!/bin/bash

# Easypanel deployment script for Job Platform
# This script prepares the environment for deployment on Easypanel

echo "Starting Job Platform deployment..."

# Check if required tools are installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "Error: Docker Compose is not installed"
    exit 1
fi

# Build Docker image
echo "Building Docker image..."
docker build -t job-platform .

# Create docker-compose.yml for Easypanel deployment
cat > docker-compose.yml << EOF
version: '3.8'
services:
  job-platform:
    image: job-platform
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgresql://user:password@db:5432/jobplatform
      - PORT=8080
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:13
    environment:
      - POSTGRES_DB=jobplatform
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  postgres_data:
EOF

echo "Docker compose configuration created."

# Create environment file template
cat > .env.template << EOF
# Environment variables for Job Platform
# Copy this file to .env and update values as needed

DATABASE_URL=postgresql://user:password@localhost:5432/jobplatform
PORT=8080
LOG_LEVEL=info
EOF

echo ".env template created."

echo "Deployment preparation complete!"
echo ""
echo "Next steps:"
echo "1. Update .env with your database connection details"
echo "2. Deploy using Docker Compose: docker-compose up -d"
echo "3. Configure Easypanel to use the Docker containers"
echo "4. Set up reverse proxy if needed"
echo ""
echo "For Easypanel-specific deployment:"
echo "- Add the Docker image to your container registry"
echo "- Configure container settings in Easypanel dashboard"
echo "- Set environment variables in Easypanel UI"
echo "- Enable auto-restart policy"