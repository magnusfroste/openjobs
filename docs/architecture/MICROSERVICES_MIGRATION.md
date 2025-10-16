# OpenJobs Microservices Architecture

**Date:** October 16, 2025  
**Status:** ✅ Complete - All Connectors Running as Standalone Microservices

## Overview

All OpenJobs connectors now run as independent microservices, enabling:
- ✅ Independent scaling per connector
- ✅ Isolated failures (one connector down doesn't affect others)
- ✅ Separate deployment cycles
- ✅ Better resource management
- ✅ Easier monitoring and debugging

## Architecture

### Before (Integrated)
```
┌─────────────────────────────────────┐
│         OpenJobs Service            │
│                                     │
│  ┌────────────────────────────┐   │
│  │  Scheduler                 │   │
│  │  - Arbetsförmedlingen      │   │
│  │  - EURES                   │   │
│  │  - Remotive                │   │
│  │  - RemoteOK  ⚠️            │   │
│  └────────────────────────────┘   │
│                                     │
│         Port 8080                   │
└─────────────────────────────────────┘
```

### After (Microservices) ✅
```
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ Arbetsförmedl.   │  │     EURES        │  │    Remotive      │  │    RemoteOK      │
│   Plugin         │  │     Plugin       │  │     Plugin       │  │     Plugin       │
│                  │  │                  │  │                  │  │                  │
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

## Migration Changes

### 1. Created Standalone Binary for RemoteOK ✅
```
cmd/plugin-remoteok/
└── main.go          # HTTP server with /health, /sync, /jobs endpoints
```

### 2. Updated Dockerfile ✅
```dockerfile
# Now builds standalone plugin
RUN go build -o plugin-remoteok ./cmd/plugin-remoteok
EXPOSE 8084
CMD ["./plugin-remoteok"]
```

### 3. Removed from Main Scheduler ✅
```go
// internal/scheduler/scheduler.go
// REMOVED: registry.Register(remoteok.NewRemoteOKConnector(store))
// Now runs as standalone microservice
```

### 4. Created Docker Compose ✅
```yaml
# docker-compose.plugins.yml
services:
  plugin-arbetsformedlingen:  # Port 8081
  plugin-eures:               # Port 8082
  plugin-remotive:            # Port 8083
  plugin-remoteok:            # Port 8084 ⭐ NEW
```

## Service Ports

| Service | Port | Purpose |
|---------|------|---------|
| **OpenJobs API** | 8080 | Main API (jobs listing, health) |
| **Arbetsförmedlingen** | 8081 | Swedish jobs connector |
| **EURES** | 8082 | European jobs connector |
| **Remotive** | 8083 | Remote jobs connector |
| **RemoteOK** | 8084 | Remote tech jobs connector |

## API Endpoints (Each Plugin)

All plugins expose the same REST API:

### GET /health
Health check endpoint
```bash
curl http://localhost:8084/health
```

Response:
```json
{
  "status": "healthy",
  "plugin": "RemoteOK Connector",
  "plugin_id": "remoteok",
  "version": "1.0.0"
}
```

### POST /sync
Trigger job synchronization
```bash
curl -X POST http://localhost:8084/sync
```

Response:
```json
{
  "success": true,
  "message": "RemoteOK sync completed successfully"
}
```

### GET /jobs
Fetch latest jobs (without storing)
```bash
curl http://localhost:8084/jobs
```

Response:
```json
{
  "success": true,
  "data": [...],
  "count": 96
}
```

## Deployment

### Option 1: Docker Compose (Recommended)

**Start all services:**
```bash
docker-compose -f docker-compose.plugins.yml up -d
```

**Start specific plugin:**
```bash
docker-compose -f docker-compose.plugins.yml up -d plugin-remoteok
```

**View logs:**
```bash
docker-compose -f docker-compose.plugins.yml logs -f plugin-remoteok
```

**Stop all:**
```bash
docker-compose -f docker-compose.plugins.yml down
```

### Option 2: Individual Containers

**Build:**
```bash
docker build -f connectors/remoteok/Dockerfile -t plugin-remoteok .
```

**Run:**
```bash
docker run -d \
  -p 8084:8084 \
  -e DATABASE_URL=$DATABASE_URL \
  --name plugin-remoteok \
  plugin-remoteok
```

### Option 3: Local Development

**Build:**
```bash
go build -o plugin-remoteok ./cmd/plugin-remoteok
```

**Run:**
```bash
PORT=8084 ./plugin-remoteok
```

## Environment Variables

Each plugin needs:
```bash
DATABASE_URL=postgresql://...    # Shared database
PORT=808X                        # Plugin port (8081-8084)
```

Additional for EURES:
```bash
ADZUNA_API_ID=your-id
ADZUNA_API_KEY=your-key
```

## Monitoring

### Health Checks
```bash
# Check all plugins
curl http://localhost:8081/health  # Arbetsförmedlingen
curl http://localhost:8082/health  # EURES
curl http://localhost:8083/health  # Remotive
curl http://localhost:8084/health  # RemoteOK
```

### Manual Sync
```bash
# Trigger sync for all
curl -X POST http://localhost:8081/sync
curl -X POST http://localhost:8082/sync
curl -X POST http://localhost:8083/sync
curl -X POST http://localhost:8084/sync
```

### Logs
```bash
# Docker Compose
docker-compose -f docker-compose.plugins.yml logs -f plugin-remoteok

# Docker
docker logs -f plugin-remoteok

# Local
tail -f /tmp/plugin-remoteok.log
```

## Benefits Achieved

### 1. Independent Scaling
```bash
# Scale RemoteOK to 3 instances
docker-compose -f docker-compose.plugins.yml up -d --scale plugin-remoteok=3
```

### 2. Isolated Failures
- If RemoteOK fails, other connectors continue working
- No cascading failures

### 3. Separate Deployments
- Deploy RemoteOK updates without touching other services
- Rollback individual plugins

### 4. Resource Management
- Allocate more resources to high-traffic connectors
- Monitor per-connector resource usage

### 5. Easier Debugging
- Isolated logs per connector
- Independent health checks
- Clear service boundaries

## Testing

### Test RemoteOK Plugin
```bash
# 1. Build
go build -o plugin-remoteok ./cmd/plugin-remoteok

# 2. Run
PORT=8084 ./plugin-remoteok &

# 3. Health check
curl http://localhost:8084/health

# 4. Sync
curl -X POST http://localhost:8084/sync

# 5. Verify jobs
curl http://localhost:8080/jobs?limit=5 | jq '.data[] | select(.fields.source == "remoteok")'
```

## Production Deployment

### Kubernetes (Future)
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: plugin-remoteok
spec:
  replicas: 2
  selector:
    matchLabels:
      app: plugin-remoteok
  template:
    metadata:
      labels:
        app: plugin-remoteok
    spec:
      containers:
      - name: plugin-remoteok
        image: openjobs/plugin-remoteok:latest
        ports:
        - containerPort: 8084
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: openjobs-secrets
              key: database-url
```

### Easypanel / Coolify
1. Create new service for each plugin
2. Set environment variables
3. Map ports (8081-8084)
4. Deploy

## Rollback Plan

If needed, you can revert to integrated mode:

1. **Uncomment in scheduler:**
```go
// internal/scheduler/scheduler.go
registry.Register(remoteok.NewRemoteOKConnector(store))
```

2. **Rebuild main service:**
```bash
docker build -t openjobs .
```

3. **Stop standalone plugin:**
```bash
docker-compose -f docker-compose.plugins.yml stop plugin-remoteok
```

## Summary

✅ **All 4 connectors now run as standalone microservices**  
✅ **RemoteOK successfully migrated from integrated to standalone**  
✅ **Docker Compose configuration ready**  
✅ **Tested and working**  

**Next Steps:**
1. Deploy to production using docker-compose.plugins.yml
2. Set up monitoring/alerting per service
3. Configure auto-scaling if needed
4. Set up cron jobs or use scheduler for periodic syncs

---

**Migration Complete!** 🎉 All connectors are now independent microservices.
