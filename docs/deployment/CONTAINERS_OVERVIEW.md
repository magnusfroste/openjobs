# OpenJobs Container Architecture

**All plugins run in separate containers** ✅

## Container Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Docker Network                               │
│                      (openjobs-network)                             │
│                                                                     │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐ │
│  │   Container 1    │  │   Container 2    │  │   Container 3    │ │
│  │ Arbetsförmedl.   │  │     EURES        │  │    Remotive      │ │
│  │     Plugin       │  │     Plugin       │  │     Plugin       │ │
│  │                  │  │                  │  │                  │ │
│  │  Port: 8081      │  │  Port: 8082      │  │  Port: 8083      │ │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘ │
│                                                                     │
│  ┌──────────────────┐  ┌──────────────────┐                       │
│  │   Container 4    │  │   Container 5    │                       │
│  │    RemoteOK      │  │  OpenJobs API    │                       │
│  │     Plugin       │  │  (Main Service)  │                       │
│  │                  │  │                  │                       │
│  │  Port: 8084      │  │  Port: 8080      │                       │
│  └──────────────────┘  └──────────────────┘                       │
│                                                                     │
│                            ↓                                        │
│                  ┌──────────────────┐                              │
│                  │  Shared Database │                              │
│                  │    (Supabase)    │                              │
│                  └──────────────────┘                              │
└─────────────────────────────────────────────────────────────────────┘
```

## Containers

| Container Name | Service | Port | Dockerfile | Binary |
|---------------|---------|------|------------|--------|
| `openjobs-api` | Main API | 8080 | `Dockerfile` | `openjobs` |
| `openjobs-plugin-arbetsformedlingen` | Swedish Jobs | 8081 | `connectors/arbetsformedlingen/Dockerfile` | `plugin-arbetsformedlingen` |
| `openjobs-plugin-eures` | European Jobs | 8082 | `connectors/eures/Dockerfile` | `plugin-eures` |
| `openjobs-plugin-remotive` | Remote Jobs | 8083 | `connectors/remotive/Dockerfile` | `plugin-remotive` |
| `openjobs-plugin-remoteok` | Remote Tech Jobs | 8084 | `connectors/remoteok/Dockerfile` | `plugin-remoteok` |

## Start All Containers

```bash
docker-compose -f docker-compose.plugins.yml up -d
```

**Output:**
```
Creating network "openjobs-network" ... done
Creating openjobs-api ... done
Creating openjobs-plugin-arbetsformedlingen ... done
Creating openjobs-plugin-eures ... done
Creating openjobs-plugin-remotive ... done
Creating openjobs-plugin-remoteok ... done
```

## Verify All Containers Running

```bash
docker-compose -f docker-compose.plugins.yml ps
```

**Expected output:**
```
NAME                                  STATUS    PORTS
openjobs-api                          Up        0.0.0.0:8080->8080/tcp
openjobs-plugin-arbetsformedlingen    Up        0.0.0.0:8081->8081/tcp
openjobs-plugin-eures                 Up        0.0.0.0:8082->8082/tcp
openjobs-plugin-remotive              Up        0.0.0.0:8083->8083/tcp
openjobs-plugin-remoteok              Up        0.0.0.0:8084->8084/tcp
```

## Health Check All Containers

```bash
# Check all plugins
curl http://localhost:8081/health | jq .plugin
curl http://localhost:8082/health | jq .plugin
curl http://localhost:8083/health | jq .plugin
curl http://localhost:8084/health | jq .plugin
```

**Expected output:**
```
"Arbetsförmedlingen Connector"
"EURES Connector"
"Remotive Connector"
"RemoteOK Connector"
```

## Container Management

### Start Individual Container
```bash
docker-compose -f docker-compose.plugins.yml up -d plugin-remoteok
```

### Stop Individual Container
```bash
docker-compose -f docker-compose.plugins.yml stop plugin-remoteok
```

### Restart Individual Container
```bash
docker-compose -f docker-compose.plugins.yml restart plugin-remoteok
```

### View Logs
```bash
# All containers
docker-compose -f docker-compose.plugins.yml logs -f

# Specific container
docker-compose -f docker-compose.plugins.yml logs -f plugin-remoteok
```

### Scale Container (if needed)
```bash
# Run 3 instances of RemoteOK
docker-compose -f docker-compose.plugins.yml up -d --scale plugin-remoteok=3
```

## Resource Allocation

Each container can have resource limits:

```yaml
plugin-remoteok:
  # ... other config
  deploy:
    resources:
      limits:
        cpus: '0.5'
        memory: 512M
      reservations:
        cpus: '0.25'
        memory: 256M
```

## Container Communication

All containers communicate via:
1. **Shared Database** - PostgreSQL (Supabase)
2. **Docker Network** - `openjobs-network`
3. **REST APIs** - Each plugin exposes HTTP endpoints

## Monitoring

### Container Stats
```bash
docker stats openjobs-plugin-remoteok
```

### Container Logs
```bash
docker logs -f openjobs-plugin-remoteok
```

### Health Status
```bash
docker inspect --format='{{.State.Health.Status}}' openjobs-plugin-remoteok
```

## Benefits of Separate Containers

### 1. Isolation ✅
- Each plugin runs independently
- Failure in one doesn't affect others
- Clean separation of concerns

### 2. Scalability ✅
- Scale each plugin independently
- Allocate resources per plugin
- Handle different load patterns

### 3. Deployment ✅
- Deploy updates per plugin
- Rollback individual plugins
- Zero downtime updates

### 4. Development ✅
- Test plugins independently
- Debug in isolation
- Faster iteration

### 5. Monitoring ✅
- Per-container metrics
- Individual logs
- Clear service boundaries

## Production Deployment

### Docker Swarm
```bash
docker stack deploy -c docker-compose.plugins.yml openjobs
```

### Kubernetes
Each plugin becomes a Deployment:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: plugin-remoteok
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: plugin-remoteok
        image: openjobs/plugin-remoteok:latest
        ports:
        - containerPort: 8084
```

### Easypanel / Coolify
1. Create 5 separate services
2. Upload docker-compose.plugins.yml
3. Deploy

## Summary

✅ **5 separate containers**:
1. Main API (8080)
2. Arbetsförmedlingen Plugin (8081)
3. EURES Plugin (8082)
4. Remotive Plugin (8083)
5. RemoteOK Plugin (8084) ⭐

✅ **Each plugin**:
- Runs in its own container
- Has its own Dockerfile
- Exposes REST API
- Connects to shared database
- Can be scaled independently

✅ **Management**:
- Start/stop individually
- View logs per container
- Monitor resources per container
- Deploy updates independently

---

**All plugins run in separate containers!** 🐳
