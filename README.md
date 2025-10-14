# OpenJobs - Open Job Listings Initiative

An open-source initiative to create a transparent, community-driven job platform where job listings are public and accessible to everyone. We believe in Open Formats, where jobs are shared freely without paywalls or walled gardens.

## Vision

OpenJobs aims to transform the job market by:
- Making job listings completely public and accessible
- Enabling innovation through open data formats
- Empowering talent matching with AI assistance
- Building a collaborative ecosystem for job seekers and employers

## Core Philosophy

**Open Access**: All job listings are publicly available
**Open Data**: Standardized formats for easy integration
**Open Innovation**: Community-driven development and improvements
**Open Talent**: AI-assisted matching of skills and opportunities

## Features

### For Employers
- **Free job posting**: Publish positions without fees
- **Transparent reach**: See who's viewing and applying
- **Community visibility**: Connect with diverse talent pools

### For Job Seekers  
- **Public job access**: Browse all listings without restrictions
- **AI-powered matching**: Smart talent connections
- **Open career resources**: Community-driven career advice

### For Developers
- **Open API**: Standardized interfaces for integration
- **Plugin ecosystem**: Extend functionality with custom connectors
- **Community collaboration**: Share improvements and innovations

## Architecture

### Container-Based Plugin System
Each data source connector runs in its own container:
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Main Platform │    │   Plugin Runner │    │   Plugin Runner │
│   (OpenJobs)    │◄──▶│   (Arbetsförmed)│◄──▶│   (LinkedIn)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Plugin DB     │    │   Data Cache    │    │   Data Cache    │
│   (Registry)    │    │   (Temp)        │    │   (Temp)        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Open API Endpoints
```
GET /health                  # System health check
GET /jobs                    # Retrieve job listings
GET /jobs/{id}               # Retrieve specific job
POST /jobs                   # Create new job listing
PUT /jobs/{id}               # Update job listing
DELETE /jobs/{id}            # Remove job listing
GET /plugins                 # List registered plugins
POST /plugins/register       # Register new plugin
GET /sync/manual             # Manual sync trigger
GET /config                  # Platform configuration
```

## Deployment

### Easypanel Deployment
OpenJobs is designed for easy deployment on Easypanel:

1. **Prerequisites**:
   - Ubuntu server with Docker installed
   - Easypanel dashboard access
   - PostgreSQL database (or external service)

2. **Deployment Steps**:
```bash
# Clone repository
git clone https://github.com/openjobs/openjobs.git
cd openjobs

# Build Docker image
docker build -t openjobs .

# Deploy to Easypanel
# Configure container settings in Easypanel dashboard
```

3. **Environment Configuration**:
```env
SYNC_FREQUENCY=3600           # Sync interval in seconds
PLUGIN_SYNC_ENABLED=true      # Enable automatic plugin sync
MAX_JOBS_PER_SYNC=100         # Maximum jobs per sync cycle
LOG_LEVEL=info                # Logging verbosity
```

## Plugin Architecture

### Plugin System
Plugins enable integration with various job sources:
- **Arbetsförmedlingen Connector**: Swedish job board integration
- **LinkedIn Jobs Connector**: Global job platform connectivity
- **Custom Connectors**: Community-developed data sources

### Plugin Development
Create plugins using the standardized interface:
1. Implement plugin entry point
2. Configure environment variables
3. Define data transformation rules
4. Register with OpenJobs platform

## Getting Started

### Prerequisites
- Go 1.19+
- Docker (for container deployment)

### Installation
```bash
# Clone the repository
git clone https://github.com/openjobs/openjobs.git
cd openjobs

# Build and run
go build -o openjobs ./cmd/openjobs
./openjobs
```

### Database Setup
OpenJobs requires PostgreSQL. Set up the database:

```bash
# Create database
createdb openjobs

# Run migrations
psql -d openjobs -f migrations/001_create_job_posts.sql
```

Set the DATABASE_URL environment variable:
```bash
export DATABASE_URL="postgresql://user:password@localhost:5432/openjobs?sslmode=disable"
```

### Docker Deployment
```bash
# Build Docker image
docker build -t openjobs .

# Run container
docker run -p 8080:8080 -e DATABASE_URL="postgresql://..." openjobs
```

## Project Structure

```
openjobs/
├── cmd/openjobs/          # Application entry point
├── pkg/
│   ├── models/           # Data models
│   └── storage/          # Database operations
├── internal/
│   ├── api/             # HTTP handlers
│   └── database/        # Database connection
├── connectors/           # Plugin connectors (future)
├── migrations/           # Database migrations
├── docs/                 # Documentation
├── Dockerfile            # Container definition
├── docker-compose.yml    # Multi-service setup
└── README.md
```

## Community & Innovation

### Open Source Mission
OpenJobs is committed to:
- **Transparency**: All job data is publicly accessible
- **Collaboration**: Community-driven platform development
- **Innovation**: AI-powered talent matching systems
- **Accessibility**: Free access for everyone

### Future Vision
- **AI Talent Matching**: Advanced algorithms for skill-persona matching
- **Global Expansion**: Support for international job markets
- **Skill Analytics**: Data-driven career development insights
- **Community Platforms**: Developer and employer forums

## Contributing

We welcome contributions to make OpenJobs better for everyone!

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, please open an issue on GitHub or contact the maintainers.

---

*Powered by OpenJobs - The Open Job Listings Initiative*