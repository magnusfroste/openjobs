# OpenJobs Microservices Architecture Guide

This document describes the **full microservices deployment** of OpenJobs, where each plugin runs in its own container with the core API separate.

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Core API      │────► Plugin AF        │    │   Shared DB     │
│   (Go)          │    │ (Go Container)   │◄──►│   (Supabase)    │
│                 │    │ - HTTP Server    │    │                 │
│ - API Endpoints │    │ - Job Sync       │    │ - job_posts     │
│ - Scheduler     │    │ - Arbetsförmedl.│    │ - plugins        │
│ - Plugin Mgmt   │    └──────────────────┘    └─────────────────┘
└─────────────────┘                                      ▲
         │                                        Shared Storage
         │
         ▼
┌─────────────────┐    ┌──────────────────┐
│ Plugin EURES    │    │ Future Python    │
│ (Go Container)  │    │ Plugin           │
│ - HTTP Server   │    │ (Python/FastAPI) │
│ - Job Sync      │    │ - HTTP Server    │
│ - Adzuna API    │    │ - ML/AI Features │
└─────────────────┘    └─────────────────┘
```

## Benefits of Microservices Architecture

### 🚀 **Independent Scaling**
- Scale Arbetsförmedlingen plugin without affecting EURES
- Handle Swedish traffic spikes separately from EU job searches

### 🐍 **Polyglot Plugins**
- Core stays in Go for performance
- Python plugins can use ML/AI libraries
- Future plugins can use JavaScript, Rust, etc.

### 🛡️ **Fault Isolation**
- Arbetsförmedlingen API down? Only that plugin fails
- Core API and other plugins remain operational

### 📦 **Independent Deployment**
- Deploy new plugin version without core downtime
- Rollback individual plugins without affecting others

## Container Architecture

### Core API Container (`openjobs-core`)
- **Purpose**: API endpoints, scheduling, plugin orchestration
- **Language**: Go 1.21
- **Port**: 8080
- **Environment**: All SUPABASE_* + PLUGIN_*_URL variables

### Plugin Containers
Each plugin runs in its own container with:
- **Shared DB Access**: All containers write to same Supabase instance
- **HTTP Interface**: Standardized `/health`, `/sync`, `/jobs` endpoints
- **Isolated Failures**: Container crashes don't affect others

## Network Communication

### Plugin Interface Contract
Plugins expose HTTP endpoints:

```javascript
GET  /health       // Plugin health status
POST /sync         // Trigger job synchronization  
GET  /jobs         // Get latest jobs fetched
```

### Core-to-Plugin Communication
```go
// Environment-based discovery
PLUGIN_ARBETSFORMEDLINGEN_URL=http://plugin-af:8081
PLUGIN_EURES_URL=http://plugin-eures:8082

// HTTP calls instead of direct method calls
resp, _ := http.Post(pluginURL + "/sync", "application/json", nil)
```

## Deployment Files

### Structure
```
openjobs/
├── Dockerfile                          // Core API container
├── Dockerfile.plugin-arbetsformedlingen  // Plugin containers
├── infrastructure.json                // Service definitions
├── deploy-microservices.sh            // Deployment script
└── cmd/
    ├── openjobs/main.go               // Core API binary
    └── plugin-arbetsformedlingen/main.go // Plugin binary
```

### Deployment Script Usage

```bash
# Deploy all services
./deploy-microservices.sh

# View logs
./deploy-microservices.sh logs

# Stop all services
./deploy-microservices.sh stop
```

### Docker Network Setup
```yaml
# infrastructure.json - Service configuration
{
  "containers": [
    {
      "name": "core-api",
      "env": ["PLUGIN_ARBETSFORMEDLINGEN_URL=http://plugin-af:8081"]
    },
    {
      "name": "plugin-arbetsformedlingen", 
      "ports": [8081]
    }
  ],
  "networks": ["openjobs-network"]
}
```

## Environment Variables

### Core API Environment
```bash
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
PLUGIN_ARBETSFORMEDLINGEN_URL=http://plugin-af:8081
PLUGIN_EURES_URL=http://plugin-eures:8082
```

### Plugin Environment
```bash
SUPABASE_URL=https://your-project.supabase.co  # Shared DB
SUPABASE_ANON_KEY=your-anon-key
ADZUNA_APP_ID=plugin-specific-key            # Plugin-specific keys
PORT=8081                                    # Container port
```

## Adding New Plugins

### Step 1: Create Plugin Binary
```go
// cmd/plugin-myclient/main.go
func main() {
    connector := myclient.NewMyClientConnector(store)
    server := &PluginServer{connector: connector}
    http.ListenAndServe(":"+port, nil)
}
```

### Step 2: Create Dockerfile
```dockerfile
# Dockerfile.plugin-myclient
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o plugin-myclient ./cmd/plugin-myclient
EXPOSE 8083
CMD ["./plugin-myclient"]
```

### Step 3: Update Environment
```bash
# In core API environment
PLUGIN_MYCLIENT_URL=http://plugin-myclient:8083
```

### Step 4: Test Deployment
```bash
# Rebuild and redeploy
docker build -t openjobs-plugin-myclient -f Dockerfile.plugin-myclient .
docker run -d --name plugin-myclient -p 8083:8083 openjobs-plugin-myclient
```

## Python Plugin Example

Here's how you could implement a Python plugin:

### main.py (FastAPI)
```python
from fastapi import FastAPI
import httpx
import os
from openjobs_shared import JobPost  # Shared types

app = FastAPI()

@app.get("/health")
async def health():
    return {"status": "healthy", "plugin": "My Python Plugin"}

@app.post("/sync")
async def sync():
    # Your Python plugin logic
    jobs = await fetch_jobs_python_style()
    return {"success": True}

@app.get("/jobs")
async def jobs():
    jobs = await fetch_jobs_python_style()
    return {"success": True, "data": jobs}
```

### Benefits for Python Plugins
- Use Python libraries for ML/AI job matching
- Leverage scikit-learn, TensorFlow for job analysis
- Access Python data science ecosystem
- Easier integration with existing Python services

## Monitoring & Debugging

### Health Checks
Each service provides `/health` endpoint:
```bash
# Check all services
curl http://localhost:8080/health  # Core
curl http://localhost:8081/health  # Plugin AF
curl http://localhost:8082/health  # Plugin EURES
```

### Service Logs
```bash
# View individual service logs
docker logs openjobs-core
docker logs openjobs-plugin-af
docker logs openjobs-plugin-eures
```

### Troubleshooting
```bash
# Check network connectivity
docker network ls
docker network inspect openjobs-network

# Test inter-service communication
curl http://plugin-af:8081/health  # From core container
```

## Migration From Single Binary

### Current State (Single Binary)
- ✅ All plugins compiled together
- ✅ Direct method calls between components
- ❌ All plugins restart when core updates

### Target State (Microservices)
- ✅ Independent plugin deployment
- ✅ Plugins communicate via HTTP
- ✅ Polyglot plugin support
- ✅ Better fault isolation

### Migration Steps
1. ✅ **Plugin Interface**: Already decoupled via interfaces
2. ✅ **HTTP Connector**: Network communication layer added
3. ⏳ **Standalone Binaries**: Plugin main.go files created
4. ⏳ **Containerization**: Individual Dockerfiles created
5. ⏳ **Deployment Script**: Orchestration added

## Next Steps

This microservices architecture is perfect for:

1. **Production Deployment**: Independent scaling and deployment
2. **Plugin Diversity**: Python, JavaScript, Rust plugins
3. **Global Expansion**: Region-specific plugin deployments
4. **Advanced Features**: ML-powered matching plugins

The foundation is now solid for building a truly distributed, scalable job platform! 🚀
