# OpenJobs Connector Architecture

## Overview

OpenJobs supports two deployment modes for connectors:

1. **Integrated Mode** - Connectors run as part of main OpenJobs service
2. **Standalone Mode** - Connectors run as separate microservices

## Current Connectors

| Connector | Mode | Dockerfile | Standalone Binary | Status |
|-----------|------|------------|-------------------|--------|
| **Arbetsförmedlingen** | Both | ✅ | ✅ `cmd/plugin-arbetsformedlingen` | Production |
| **EURES** | Both | ✅ | ✅ `cmd/plugin-eures` | Production |
| **Remotive** | Both | ✅ | ✅ `cmd/plugin-remotive` | Production |
| **RemoteOK** | Integrated | ✅ | ❌ (runs in main) | Production |

## Architecture Patterns

### Pattern 1: Standalone Plugin (Arbetsförmedlingen, EURES, Remotive)

```
connectors/
└── arbetsformedlingen/
    ├── connector.go          # Connector implementation
    ├── Dockerfile            # Standalone deployment
    └── README.md

cmd/
└── plugin-arbetsformedlingen/
    └── main.go               # Standalone entry point
```

**Benefits:**
- Can run as separate microservice
- Independent scaling
- Isolated failures
- Separate deployment

**Usage:**
```bash
# Build standalone
docker build -f connectors/arbetsformedlingen/Dockerfile -t plugin-af .

# Run standalone
docker run -p 8081:8081 plugin-af
```

### Pattern 2: Integrated Connector (RemoteOK)

```
connectors/
└── remoteok/
    ├── connector.go          # Connector implementation
    ├── Dockerfile            # For consistency (builds main service)
    └── README.md

internal/scheduler/
└── scheduler.go              # Registers connector
```

**Benefits:**
- Simpler deployment (one service)
- Lower resource usage
- Easier local development
- Shared database connection

**Usage:**
```go
// Automatically registered in main service
registry.Register(remoteok.NewRemoteOKConnector(store))
```

## When to Use Each Pattern

### Use Standalone Pattern When:
- ✅ Connector needs independent scaling
- ✅ Connector has high resource usage
- ✅ Connector needs separate deployment cycle
- ✅ Connector may fail independently
- ✅ External team maintains connector

### Use Integrated Pattern When:
- ✅ Connector is lightweight
- ✅ Connector shares resources with main service
- ✅ Simpler deployment preferred
- ✅ Low traffic/usage
- ✅ Core team maintains connector

## Connector Interface

All connectors implement the same interface:

```go
type PluginConnector interface {
    GetID() string
    GetName() string
    FetchJobs() ([]models.JobPost, error)
    SyncJobs() error
}
```

## Registration

### Integrated Mode
```go
// internal/scheduler/scheduler.go
func NewScheduler(store *storage.JobStore) *Scheduler {
    registry := NewPluginRegistry()
    
    // Register integrated connectors
    registry.Register(remoteok.NewRemoteOKConnector(store))
    
    return &Scheduler{registry: registry}
}
```

### Standalone Mode
```go
// cmd/plugin-arbetsformedlingen/main.go
func main() {
    store := storage.NewJobStore(db)
    connector := arbetsformedlingen.NewArbetsformedlingenConnector(store)
    
    // Run as HTTP service
    http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
        connector.SyncJobs()
    })
    
    http.ListenAndServe(":8081", nil)
}
```

## Migration Path: Integrated → Standalone

To convert RemoteOK to standalone:

### Step 1: Create Standalone Binary
```go
// cmd/plugin-remoteok/main.go
package main

import (
    "openjobs/connectors/remoteok"
    "openjobs/pkg/storage"
)

func main() {
    // Initialize
    db := initDatabase()
    store := storage.NewJobStore(db)
    connector := remoteok.NewRemoteOKConnector(store)
    
    // Run scheduler
    scheduler := NewPluginScheduler(connector)
    scheduler.Start()
}
```

### Step 2: Update Dockerfile
```dockerfile
# Build standalone binary
RUN CGO_ENABLED=0 GOOS=linux go build -o plugin-remoteok ./cmd/plugin-remoteok

# Copy binary
COPY --from=builder /app/plugin-remoteok .

# Run standalone
CMD ["./plugin-remoteok"]
```

### Step 3: Update Scheduler
```go
// Remove from internal/scheduler/scheduler.go
// registry.Register(remoteok.NewRemoteOKConnector(store))
```

## Current Deployment

### Main OpenJobs Service
```yaml
# docker-compose.yml
services:
  openjobs:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=${DATABASE_URL}
    # Includes: RemoteOK (integrated)
```

### Standalone Plugins
```yaml
# docker-compose.yml
services:
  plugin-arbetsformedlingen:
    build:
      context: .
      dockerfile: connectors/arbetsformedlingen/Dockerfile
    ports:
      - "8081:8081"
      
  plugin-eures:
    build:
      context: .
      dockerfile: connectors/eures/Dockerfile
    ports:
      - "8082:8082"
      
  plugin-remotive:
    build:
      context: .
      dockerfile: connectors/remotive/Dockerfile
    ports:
      - "8083:8083"
```

## Recommendation

**Current setup is good!** 

- **RemoteOK** = Integrated (lightweight, simple API)
- **Others** = Standalone (more complex, independent scaling)

**Keep RemoteOK integrated** unless:
- You need independent scaling
- You want to deploy it separately
- You have resource constraints

## Summary

✅ **RemoteOK Dockerfile created** for consistency  
✅ **Currently runs integrated** in main service  
✅ **Can be converted to standalone** if needed  
✅ **All connectors follow same interface**  

The architecture is flexible - you can run connectors integrated or standalone based on your needs!
