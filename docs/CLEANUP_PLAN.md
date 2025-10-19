# OpenJobs Cleanup Plan

## 📋 Files to Remove

### Obsolete Documentation (Root Level)
- [ ] `CHANGELOG_FIXES.md` - Merge into main README or delete
- [ ] `DASHBOARD_FIXES.md` - Outdated troubleshooting
- [ ] `INDEX.md` - Redundant with README
- [ ] `REMOTEOK_INTEGRATION.md` - Now part of standard connectors
- [ ] `REMOTIVE_SSL_ISSUE.md` - Handled in code
- [ ] `SYNC_LOGGING_FIX.md` - Already implemented
- [ ] `SYNC_LOGGING_GUIDE.md` - Merge into main docs
- [ ] `VERSION_GUIDE.md` - Outdated

### Obsolete Binaries
- [ ] `arbetsformedlingen-connector` - Old binary
- [ ] `openjobs` - Old binary
- [ ] `plugin-remoteok` - Old binary
- [ ] `bin/` - Contains old binaries

### Obsolete Scripts
- [ ] `test.tar.gz` - Test artifact
- [ ] `test_local.sh` - Outdated testing script
- [ ] `build.sh` - Redundant with Docker
- [ ] `run-local.sh` - Redundant with Docker

### Keep & Update
- ✅ `README.md` - Main documentation (update)
- ✅ `QUICKSTART.md` - Quick start guide (update)
- ✅ `SETUP_GUIDE.md` - Setup instructions (update)
- ✅ `EASYPANEL_ENV_SETUP.md` - Deployment guide (keep)
- ✅ `docs/` - Consolidated documentation
- ✅ `.env.example` - Environment template
- ✅ `Dockerfile` - Container build
- ✅ `docker-compose.plugins.yml` - Local development
- ✅ `deploy-microservices.sh` - Deployment script

## 📁 New Documentation Structure

```
/Users/mafr/Code/OpenJobs/
├── README.md                          # Main overview
├── QUICKSTART.md                      # 5-minute setup
├── docs/
│   ├── README.md                      # Docs index
│   ├── setup/
│   │   ├── local-development.md       # Docker Compose setup
│   │   ├── easypanel-deployment.md    # Easypanel guide
│   │   └── environment-variables.md   # All env vars
│   ├── architecture/
│   │   ├── overview.md                # System architecture
│   │   ├── connectors.md              # Connector design
│   │   └── incremental-sync.md        # Sync mechanism
│   ├── connectors/
│   │   ├── arbetsformedlingen.md      # AF connector
│   │   ├── eures.md                   # EURES/Adzuna
│   │   ├── remotive.md                # Remotive
│   │   └── remoteok.md                # RemoteOK
│   └── api/
│       ├── endpoints.md               # API reference
│       └── plugin-protocol.md         # HTTP plugin spec
```

## 🔧 Code Cleanup

### Remove Unused Code
- [ ] Old sync mechanisms (if any file-based state)
- [ ] Deprecated environment variable handlers
- [ ] Unused utility functions

### Consolidate
- [ ] Move all connector docs to `docs/connectors/`
- [ ] Consolidate deployment guides
- [ ] Update all references to new structure
