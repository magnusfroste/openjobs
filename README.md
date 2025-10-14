# OpenJobs - Open Job Listings Initiative

An open-source initiative to create a transparent, community-driven job platform where job listings are public and accessible to everyone. We believe in Open Formats, where jobs are shared freely without paywalls or walled gardens.

## Vision

OpenJobs aims to transform the job market by:
- Making job listings completely public and accessible
- Enabling innovation through open data formats
- Empowering talent matching with AI assistance
- Building a collaborative ecosystem for job seekers and employers

## Core Philosophy

**Open Access**: All job listings are publicly available without walled gardens
**Open Data**: Standardized formats for easy integration and sharing
**Open Innovation**: Community-driven development focused on transparency
**Open Talent**: AI-assisted matching that benefits everyone, not corporations
**Open Sharing**: We integrate with platforms that share data openly for the greater good

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
Each data source connector runs in its own container, promoting truly open data sharing:
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Main Platform │    │   Plugin Runner │    │   Plugin Runner │
│   (OpenJobs)    │◄──▶│   (Arbetsförmed)│◄──▶│   (Adzuna)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Plugin DB     │    │   Data Cache    │    │   Data Cache    │
│   (Registry)    │    │   (Open API)    │    │   (Open API)    │
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
POST /sync/manual            # Manual job data synchronization
GET /plugins                 # List registered plugins
POST /plugins/register       # Register new plugin
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
Plugins enable integration with truly open job sources that share data for the greater good:
- **Arbetsförmedlingen Connector**: Swedish public employment service (government open data)
- **Adzuna Jobs Connector**: Global job search API with generous free tier
- **Reed.co.uk Connector**: UK job board with open API access
- **EURES Connector**: European Commission job mobility portal (pan-European)
- **Authentic Jobs Connector**: Independent developer-focused job board
- **We Work Remotely Connector**: Remote work job board with open RSS feeds
- **Community Job Boards**: Local and niche platforms with open APIs
- **Government Job Portals**: Public sector employment services
- **Company Career Pages**: Direct company job posting integration
- **Custom Connectors**: Community-developed open data sources

### Plugin Development
Create plugins using the standardized interface:
1. Implement plugin entry point
2. Configure environment variables
3. Define data transformation rules
4. Register with OpenJobs platform

## Automated Data Ingestion

OpenJobs automatically ingests job data from connected sources:

### Scheduled Ingestion
- **Frequency**: Every 6 hours
- **Sources**: EURES (European job mobility portal)
- **Process**: Automatic fetching, transformation, and storage

### Manual Sync
Trigger immediate data ingestion:
```bash
curl -X POST http://localhost:8080/sync/manual
```

### Data Sources
- **EURES**: European Commission job mobility portal
- **Future**: Additional open job platforms

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
OpenJobs uses Supabase (PostgreSQL with advanced features). Set up your database:

**Option 1: Supabase Cloud (Recommended)**
1. Create account at https://supabase.com
2. Create new project
3. Go to SQL Editor and run: `migrations/001_create_job_posts.sql`

**Option 2: Self-hosted Supabase (Your Setup)**
1. Access your Supabase dashboard
2. Go to SQL Editor
3. Run the contents of `migrations/001_create_job_posts.sql`

Set the environment variables. You can either:

**Option 1: Use .env file (Recommended)**
```bash
# Copy the sample .env file and edit with your values
cp .env.example .env
# Edit .env with your Supabase credentials
```

**Option 2: Export environment variables**
```bash
export SUPABASE_URL="https://supabase.froste.eu"
export SUPABASE_ANON_KEY="your-anon-key-here"
export PORT=8080
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