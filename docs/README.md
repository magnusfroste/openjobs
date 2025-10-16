# OpenJobs Documentation

**Complete documentation for OpenJobs job aggregation platform**

## 📚 Documentation Structure

### Architecture
- [Connector Architecture](architecture/CONNECTOR_ARCHITECTURE.md) - Plugin system design patterns
- [Microservices Migration](architecture/MICROSERVICES_MIGRATION.md) - Transition to microservices

### Deployment
- [Containers Overview](deployment/CONTAINERS_OVERVIEW.md) - Docker container architecture
- [Docker Compose Guide](deployment/docker-compose.md) - Multi-container deployment

### Connectors
- [Arbetsförmedlingen](../connectors/arbetsformedlingen/README.md) - Swedish jobs
- [EURES](../connectors/eures/README.md) - European jobs
- [Remotive](../connectors/remotive/README.md) - Remote jobs
- [RemoteOK](../connectors/remoteok/README.md) - Remote tech jobs

### Migrations
- [001 - Create Job Posts](../migrations/001_create_job_posts.sql) - Initial schema
- [002 - Add Matching Fields](../migrations/002_add_matching_fields.sql) - Enhanced fields for matching

## 🚀 Quick Start

### Run All Services
```bash
docker-compose -f docker-compose.plugins.yml up -d
```

### Check Health
```bash
curl http://localhost:8080/health  # Main API
curl http://localhost:8081/health  # Arbetsförmedlingen
curl http://localhost:8082/health  # EURES
curl http://localhost:8083/health  # Remotive
curl http://localhost:8084/health  # RemoteOK
```

### Trigger Sync
```bash
curl -X POST http://localhost:8081/sync  # Arbetsförmedlingen
curl -X POST http://localhost:8082/sync  # EURES
curl -X POST http://localhost:8083/sync  # Remotive
curl -X POST http://localhost:8084/sync  # RemoteOK
```

## 🏗️ Architecture Overview

```
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ Arbetsförmedl.   │  │     EURES        │  │    Remotive      │  │    RemoteOK      │
│   Plugin         │  │     Plugin       │  │     Plugin       │  │     Plugin       │
│   Port 8081      │  │   Port 8082      │  │   Port 8083      │  │   Port 8084      │
└──────────────────┘  └──────────────────┘  └──────────────────┘  └──────────────────┘
         │                     │                     │                     │
         └─────────────────────┴─────────────────────┴─────────────────────┘
                                        │
                              ┌─────────────────────┐
                              │  Shared Database    │
                              │   (Supabase)        │
                              └─────────────────────┘
```

## 📖 Key Concepts

### Microservices Architecture
All connectors run as independent microservices:
- **Independent scaling** - Scale each connector separately
- **Isolated failures** - One connector down doesn't affect others
- **Separate deployments** - Deploy updates independently
- **Better monitoring** - Per-service logs and metrics

### Plugin Interface
All connectors implement the same interface:
```go
type PluginConnector interface {
    GetID() string
    GetName() string
    FetchJobs() ([]models.JobPost, error)
    SyncJobs() error
}
```

### Data Model
Enhanced schema for better job matching:
- `salary_min`, `salary_max`, `salary_currency` - Structured salary data
- `is_remote` - Remote work flag
- `url` - Direct application link
- `requirements[]` - Skills array
- `benefits[]` - Benefits array
- `fields` - JSONB for connector-specific data

## 🔧 Development

### Adding a New Connector

1. **Create connector directory:**
```bash
mkdir -p connectors/mynewconnector
```

2. **Implement connector:**
```go
// connectors/mynewconnector/connector.go
type MyConnector struct {
    store *storage.JobStore
}

func (c *MyConnector) GetID() string { return "mynewconnector" }
func (c *MyConnector) GetName() string { return "My New Connector" }
func (c *MyConnector) FetchJobs() ([]models.JobPost, error) { /* ... */ }
func (c *MyConnector) SyncJobs() error { /* ... */ }
```

3. **Create standalone binary:**
```go
// cmd/plugin-mynewconnector/main.go
func main() {
    store := storage.NewJobStore()
    connector := mynewconnector.NewMyConnector(store)
    // ... HTTP server setup
}
```

4. **Create Dockerfile:**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o plugin-mynewconnector ./cmd/plugin-mynewconnector

FROM alpine:latest
COPY --from=builder /app/plugin-mynewconnector .
EXPOSE 8085
CMD ["./plugin-mynewconnector"]
```

5. **Add to docker-compose:**
```yaml
plugin-mynewconnector:
  build:
    context: .
    dockerfile: connectors/mynewconnector/Dockerfile
  ports:
    - "8085:8085"
  environment:
    - DATABASE_URL=${DATABASE_URL}
```

## 📊 Monitoring

### Logs
```bash
# All services
docker-compose -f docker-compose.plugins.yml logs -f

# Specific service
docker-compose -f docker-compose.plugins.yml logs -f plugin-remoteok
```

### Metrics
```bash
# Container stats
docker stats openjobs-plugin-remoteok

# Health checks
watch -n 5 'curl -s http://localhost:8084/health | jq .'
```

## 🤝 Integration

### LazyJobs Integration
OpenJobs serves as the data source for LazyJobs:
1. OpenJobs aggregates jobs from 4 sources
2. LazyJobs fetches via OpenJobs API
3. AI enrichment adds skills
4. Jobs stored in LazyJobs for matching

See [LazyJobs Integration Guide](../../LazyJobs/docs/OPENJOBS_INTEGRATION.md)

## 📝 Recent Updates

### October 16, 2025
- ✅ Migrated all connectors to microservices architecture
- ✅ Added RemoteOK standalone plugin
- ✅ Enhanced data model with matching fields
- ✅ Created comprehensive documentation

## 🔗 Links

- [Main README](../README.md)
- [Docker Compose File](../docker-compose.plugins.yml)
- [Migrations](../migrations/)
- [Connectors](../connectors/)

---

**For questions or issues, please open a GitHub issue.**
