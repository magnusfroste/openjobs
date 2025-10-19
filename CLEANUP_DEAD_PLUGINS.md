# Cleanup Dead Plugins

## Plugins to Remove

### 1. Indeed API Plugin (Port 8085)
**Reason:** Indeed discontinued their public API. This plugin only returns demo data.

**Steps:**
```bash
# 1. Remove from scheduler
# Edit internal/scheduler/scheduler.go
# Remove "indeed" from pluginURLs and pluginNames

# 2. Stop container (if running)
docker stop openjobs-indeed

# 3. Archive code (don't delete - keep for reference)
mkdir -p archive/deprecated-plugins
mv connectors/indeed archive/deprecated-plugins/
mv cmd/plugin-indeed archive/deprecated-plugins/

# 4. Update documentation
# Remove from README, CONNECTORS_SUMMARY.md
```

### 2. Indeed-Scraper Plugin (Port 8086)
**Reason:** Blocked by Cloudflare with 403 errors. 0% success rate.

**Steps:**
```bash
# 1. Remove from scheduler (if added)
# Check internal/scheduler/scheduler.go

# 2. Stop container (if running)
docker stop openjobs-indeed-scraper

# 3. Archive code
mv connectors/indeed-scraper archive/deprecated-plugins/
mv cmd/plugin-indeed-scraper archive/deprecated-plugins/

# 4. Update documentation
```

## Active Plugins After Cleanup

1. **arbetsformedlingen** (8081) - Swedish Employment Service
2. **eures** (8082) - European Job Mobility (Adzuna)
3. **remotive** (8083) - Remote-first jobs
4. **remoteok** (8084) - Remote tech jobs
5. **indeed-chrome** (8087) - Indeed Sweden (headless Chrome)

## Benefits

- ✅ Cleaner codebase
- ✅ No wasted resources on dead plugins
- ✅ Clearer documentation
- ✅ Easier maintenance
- ✅ Code preserved in archive for reference

## Scheduler Update

**File:** `internal/scheduler/scheduler.go`

**Remove these lines:**
```go
// Remove from pluginURLs map
"indeed": os.Getenv("PLUGIN_INDEED_URL"),

// Remove from default URLs
if pluginURLs["indeed"] == "" {
    pluginURLs["indeed"] = "http://localhost:8085"
}

// Remove from pluginNames map
"indeed": "Indeed",
```

**Keep these:**
```go
"arbetsformedlingen": "Arbetsförmedlingen",
"eures":              "EURES",
"remotive":           "Remotive",
"remoteok":           "RemoteOK",
"indeed-chrome":      "Indeed Chrome",
```
