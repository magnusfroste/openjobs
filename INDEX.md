# OpenJobs Project Index

**Quick reference for all project files and documentation**

## 📁 Project Structure

```
OpenJobs/
├── cmd/                          # Binaries
│   ├── openjobs/                 # Main API service
│   ├── plugin-arbetsformedlingen/ # Swedish jobs plugin
│   ├── plugin-eures/             # European jobs plugin
│   ├── plugin-remotive/          # Remote jobs plugin
│   └── plugin-remoteok/          # Remote tech jobs plugin
│
├── connectors/                   # Plugin connectors
│   ├── arbetsformedlingen/       # Swedish employment service
│   ├── eures/                    # European job mobility
│   ├── remotive/                 # Remote-first platform
│   └── remoteok/                 # Remote tech jobs
│
├── pkg/                          # Shared packages
│   ├── models/                   # Data models & interfaces
│   └── storage/                  # Database operations
│
├── internal/                     # Private packages
│   ├── api/                      # HTTP handlers
│   ├── database/                 # Database connection
│   └── scheduler/                # Job scheduling
│
├── migrations/                   # Database migrations
│   ├── 001_create_job_posts.sql  # Initial schema
│   └── 002_add_matching_fields.sql # Enhanced fields
│
├── docs/                         # Documentation
│   ├── README.md                 # Documentation index
│   ├── architecture/             # Architecture docs
│   ├── deployment/               # Deployment guides
│   ├── connectors/               # Connector docs
│   └── migrations/               # Migration guides
│
├── docker-compose.plugins.yml    # Multi-container deployment
├── Dockerfile                    # Main API container
└── README.md                     # Project overview
```

## 📚 Key Documentation

### Getting Started
- [README.md](README.md) - Project overview
- [docs/README.md](docs/README.md) - Documentation hub

### Architecture
- [Connector Architecture](docs/architecture/CONNECTOR_ARCHITECTURE.md) - Plugin patterns
- [Microservices Migration](docs/architecture/MICROSERVICES_MIGRATION.md) - Migration guide
- [Containers Overview](docs/deployment/CONTAINERS_OVERVIEW.md) - Docker architecture

### Deployment
- [docker-compose.plugins.yml](docker-compose.plugins.yml) - All services
- Individual Dockerfiles in `connectors/*/Dockerfile`

### Connectors
- [Arbetsförmedlingen](connectors/arbetsformedlingen/README.md) - Swedish jobs
- [EURES](connectors/eures/README.md) - European jobs
- [Remotive](connectors/remotive/README.md) - Remote jobs
- [RemoteOK](connectors/remoteok/README.md) - Remote tech jobs

### Database
- [001 - Initial Schema](migrations/001_create_job_posts.sql)
- [002 - Matching Fields](migrations/002_add_matching_fields.sql)

## 🚀 Quick Commands

### Development
```bash
# Build main API
go build -o openjobs ./cmd/openjobs

# Build specific plugin
go build -o plugin-remoteok ./cmd/plugin-remoteok

# Run locally
./openjobs
PORT=8084 ./plugin-remoteok
```

### Docker
```bash
# Start all services
docker-compose -f docker-compose.plugins.yml up -d

# Build specific plugin
docker build -f connectors/remoteok/Dockerfile -t plugin-remoteok .

# View logs
docker-compose -f docker-compose.plugins.yml logs -f plugin-remoteok
```

### Testing
```bash
# Health checks
curl http://localhost:8080/health  # Main API
curl http://localhost:8081/health  # Arbetsförmedlingen
curl http://localhost:8082/health  # EURES
curl http://localhost:8083/health  # Remotive
curl http://localhost:8084/health  # RemoteOK

# Trigger sync
curl -X POST http://localhost:8084/sync

# Get jobs
curl http://localhost:8080/jobs?limit=10
```

## 🔧 Configuration Files

- `.env.example` - Environment variables template
- `go.mod`, `go.sum` - Go dependencies
- `docker-compose.plugins.yml` - Multi-container setup

## 📊 Service Ports

| Service | Port | Purpose |
|---------|------|---------|
| Main API | 8080 | Job listings, health |
| Arbetsförmedlingen | 8081 | Swedish jobs |
| EURES | 8082 | European jobs |
| Remotive | 8083 | Remote jobs |
| RemoteOK | 8084 | Remote tech jobs |

## 🔗 Related Projects

- **LazyJobs** - `/Users/mafr/Code/LazyJobs` - Job matching platform (consumer)
- Uses OpenJobs as primary data source

## 📝 Recent Changes (Oct 16, 2025)

- ✅ Migrated all connectors to microservices
- ✅ Added RemoteOK standalone plugin
- ✅ Enhanced data model with matching fields
- ✅ Organized documentation structure

---

**For detailed information, see [docs/README.md](docs/README.md)**
