# OpenJobs - Job Aggregation Platform

**Microservices-based job aggregation platform with intelligent incremental sync**

OpenJobs aggregates job listings from multiple sources into a unified API, featuring:
- 🔄 **Incremental Sync** - Only fetch new jobs, avoid duplicates
- 🐳 **Microservices Architecture** - Each connector runs independently
- ⏰ **Cron Scheduling** - Precise daily sync at 6:00 AM
- 📊 **Database-backed State** - No file-based persistence needed
- 🚀 **High Limits** - Fetch up to 500 jobs per connector

## 🎯 Current Status

**Production Deployment:** https://app-openjobs.katsu6.easypanel.host

| Connector | Jobs | Sync Method | Limit |
|-----------|------|-------------|-------|
| **Arbetsförmedlingen** | 50+ | API date filter | 500 |
| **EURES (Adzuna)** | 1+ | API date filter | 100 |
| **Remotive** | 100+ | Client filter | 100 |
| **RemoteOK** | 168+ | Client filter | All |
| **Total** | **333+** | Daily at 6 AM | - |

## 🏗️ Architecture

### Microservices Design
```
┌─────────────────────────────────────────────────────────┐
│                    Main API (Port 8080)                 │
│  - REST API                                             │
│  - Dashboard                                            │
│  - Scheduler (Cron)                                     │
│  - HTTP Plugin Orchestrator                             │
└────────────┬────────────────────────────────────────────┘
             │
             │ HTTP POST /sync
             │
    ┌────────┴────────┬──────────┬──────────┐
    │                 │          │          │
    ▼                 ▼          ▼          ▼
┌────────┐      ┌────────┐  ┌────────┐  ┌────────┐
│  AF    │      │ EURES  │  │Remotive│  │RemoteOK│
│ :8081  │      │ :8082  │  │ :8083  │  │ :8084  │
└────┬───┘      └────┬───┘  └────┬───┘  └────┬───┘
     │               │           │           │
     └───────────────┴───────────┴───────────┘
                     │
                     ▼
            ┌─────────────────┐
            │  Supabase DB    │
            │  - job_posts    │
            │  - sync_logs    │
            └─────────────────┘
```

## ✨ Key Features

### Intelligent Incremental Sync
- **Database-backed state**: Tracks last sync via `posted_date` in database
- **API date filtering**: Arbetsförmedlingen & EURES use API parameters
- **Client-side filtering**: Remotive & RemoteOK filter locally
- **Zero duplicates**: All syncs show 0 new when no updates

### Cron-Based Scheduling
```bash
CRON_SCHEDULE=0 6 * * *  # Daily at 6:00 AM
```
- Precise timing (not interval-based)
- Configurable per environment
- Fallback to interval mode if not set

### High Performance
- **Arbetsförmedlingen**: 500 jobs/sync (up from 20)
- **EURES**: 100 jobs/sync (up from 10)
- **Remotive**: 100 jobs/sync (up from 10)
- **RemoteOK**: All jobs (client-filtered)

### Production Ready
- Docker containerized
- Easypanel deployment
- Health checks on all services
- Comprehensive logging

## 📡 API Endpoints

### Main API (Port 8080)
```bash
# Health & Status
GET  /health                 # System health check
GET  /                       # Dashboard UI

# Jobs
GET  /jobs                   # List all jobs
GET  /jobs/:id               # Get specific job

# Sync
POST /sync/manual            # Trigger manual sync
GET  /sync/history           # View sync logs

# Plugins
GET  /plugins                # List registered plugins
```

### Plugin Endpoints (Ports 8081-8084)
```bash
GET  /health                 # Plugin health check
POST /sync                   # Trigger plugin sync
```

## 🚀 Deployment

### Quick Start (Easypanel)

**1. Deploy Main API**
```bash
Image: ghcr.io/magnusfroste/openjobs:latest
Port: 8080
```

**Environment Variables:**
```bash
SUPABASE_URL=https://supabase.froste.eu
SUPABASE_ANON_KEY=your-key-here
USE_HTTP_PLUGINS=true
CRON_SCHEDULE=0 6 * * *
PLUGIN_ARBETSFORMEDLINGEN_URL=http://plugin-arbetsformedlingen:8081
PLUGIN_EURES_URL=http://plugin-eures:8082
PLUGIN_REMOTIVE_URL=http://plugin-remotive:8083
PLUGIN_REMOTEOK_URL=http://plugin-remoteok:8084
```

**2. Deploy Each Plugin**

Create 4 services with these images:
- `ghcr.io/magnusfroste/openjobs-arbetsformedlingen:latest` (Port 8081)
- `ghcr.io/magnusfroste/openjobs-eures:latest` (Port 8082)
- `ghcr.io/magnusfroste/openjobs-remotive:latest` (Port 8083)
- `ghcr.io/magnusfroste/openjobs-remoteok:latest` (Port 8084)

Each needs:
```bash
SUPABASE_URL=https://supabase.froste.eu
SUPABASE_ANON_KEY=your-key-here
PORT=808X  # Respective port
```

**EURES also needs:**
```bash
ADZUNA_APP_ID=your-adzuna-id
ADZUNA_APP_KEY=your-adzuna-key
```

See [EASYPANEL_ENV_SETUP.md](EASYPANEL_ENV_SETUP.md) for detailed instructions.

## 🔌 Connectors

### Active Connectors

| Connector | Source | Type | Jobs |
|-----------|--------|------|------|
| **Arbetsförmedlingen** | Swedish Employment Service | Government | 50+ |
| **EURES** | Adzuna API (European jobs) | Commercial | 1+ |
| **Remotive** | Remotive.com | Platform | 100+ |
| **RemoteOK** | RemoteOK.com | Platform | 168+ |

### Connector Interface

All connectors implement:
```go
type PluginConnector interface {
    GetID() string
    GetName() string
    FetchJobs() ([]JobPost, error)
    SyncJobs() error
}
```

### Adding New Connectors

1. Create connector in `connectors/yourname/`
2. Implement `PluginConnector` interface
3. Add Dockerfile
4. Register in main scheduler
5. Deploy as new microservice

See existing connectors for examples.

## 🔄 Data Sync

### Automatic Sync
- **Schedule**: Daily at 6:00 AM (configurable via `CRON_SCHEDULE`)
- **Method**: HTTP POST to each plugin container
- **Logging**: All syncs logged to `sync_logs` table

### Manual Sync
```bash
curl -X POST https://app-openjobs.katsu6.easypanel.host/sync/manual
```

### Incremental Sync Logic

**API Date Filtering** (Arbetsförmedlingen, EURES):
1. Query database for most recent job's `posted_date`
2. Add `?published-after=YYYY-MM-DD` to API request
3. API returns only new jobs

**Client-Side Filtering** (Remotive, RemoteOK):
1. Fetch all jobs from API
2. Query database for most recent job's `posted_date`
3. Filter locally to only process new jobs
4. Skip transformation/insertion of duplicates

## 🛠️ Local Development

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Supabase account (or self-hosted)

### Quick Start

**1. Clone & Setup**
```bash
git clone https://github.com/magnusfroste/openjobs.git
cd openjobs
cp .env.example .env
# Edit .env with your Supabase credentials
```

**2. Run Database Migrations**
```sql
-- In Supabase SQL Editor, run:
migrations/001_create_job_posts.sql
migrations/002_add_job_fields.sql
```

**3. Start All Services**
```bash
docker-compose -f docker-compose.plugins.yml up
```

This starts:
- Main API: http://localhost:8080
- Arbetsförmedlingen: http://localhost:8081
- EURES: http://localhost:8082
- Remotive: http://localhost:8083
- RemoteOK: http://localhost:8084

**4. Trigger Sync**
```bash
curl -X POST http://localhost:8080/sync/manual
```

## 📁 Project Structure

```
openjobs/
├── cmd/
│   ├── openjobs/                 # Main API
│   ├── plugin-arbetsformedlingen/ # AF plugin
│   ├── plugin-eures/             # EURES plugin
│   ├── plugin-remotive/          # Remotive plugin
│   └── plugin-remoteok/          # RemoteOK plugin
├── connectors/
│   ├── arbetsformedlingen/       # AF connector logic
│   ├── eures/                    # EURES connector logic
│   ├── remotive/                 # Remotive connector logic
│   └── remoteok/                 # RemoteOK connector logic
├── pkg/
│   ├── models/                   # Data models
│   └── storage/                  # Database operations
├── internal/
│   ├── api/                      # HTTP handlers
│   └── scheduler/                # Cron scheduler
├── migrations/                   # Database migrations
├── docs/                         # Documentation
├── Dockerfile                    # Main API container
├── docker-compose.plugins.yml    # All services
└── EASYPANEL_ENV_SETUP.md        # Deployment guide
```

## 📚 Documentation

- [QUICKSTART.md](QUICKSTART.md) - 5-minute setup guide
- [SETUP_GUIDE.md](SETUP_GUIDE.md) - Detailed setup instructions
- [EASYPANEL_ENV_SETUP.md](EASYPANEL_ENV_SETUP.md) - Production deployment
- [docs/](docs/) - Architecture and API documentation

## 🤝 Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📝 License

MIT License - see LICENSE file for details.

---

**Built with ❤️ by [@magnusfroste](https://github.com/magnusfroste)**
