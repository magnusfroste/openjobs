# OpenJobs Plugin Scheduling Investigation

**Date:** October 25, 2025  
**Issue:** Plugins may not be triggering at the expected 6:00 AM schedule

---

## 🔍 Current Configuration

### Scheduler Setup (Code)

**File:** `/internal/scheduler/scheduler.go`

The scheduler supports two modes:

1. **CRON Mode** (Priority) - Uses `CRON_SCHEDULE` env var
2. **Interval Mode** (Fallback) - Uses `SYNC_INTERVAL_HOURS` env var (default: 24 hours)

```go
// Line 40-53
cronSchedule := os.Getenv("CRON_SCHEDULE")
syncIntervalHours := 24 // Default to once per day
if envInterval := os.Getenv("SYNC_INTERVAL_HOURS"); envInterval != "" {
    if hours, err := strconv.Atoi(envInterval); err == nil {
        syncIntervalHours = hours
    }
}
```

**Expected Schedule:** `CRON_SCHEDULE=0 6 * * *` (Every day at 6:00 AM)

---

## 🔌 Plugin Architecture

### Microservices Mode

The system uses **HTTP plugin containers** when `USE_HTTP_PLUGINS=true`:

```go
// Line 123-128
useHTTPPlugins := os.Getenv("USE_HTTP_PLUGINS") == "true"

if useHTTPPlugins {
    // Microservices mode: Call plugin containers via HTTP
    fmt.Println("🔌 Using HTTP plugin containers (microservices mode)")
    s.RunManualSync()
}
```

### Registered Plugins

**File:** `/internal/scheduler/scheduler.go` (Line 152-159)

```go
pluginURLs := map[string]string{
    "arbetsformedlingen": os.Getenv("PLUGIN_ARBETSFORMEDLINGEN_URL"),
    "eures":              os.Getenv("PLUGIN_EURES_URL"),
    "remotive":           os.Getenv("PLUGIN_REMOTIVE_URL"),
    "remoteok":           os.Getenv("PLUGIN_REMOTEOK_URL"),
    "indeed-chrome":      os.Getenv("PLUGIN_INDEED_CHROME_URL"),
    "jooble":             os.Getenv("PLUGIN_JOOBLE_URL"),
}
```

---

## ⚠️ Potential Issues

### 1. Missing Environment Variables in Production

**Local .env file** (development only):
```bash
SUPABASE_URL=https://supabase.froste.eu
SUPABASE_ANON_KEY=eyJ...
PORT=8080
ADZUNA_APP_ID=8d597455
ADZUNA_APP_KEY=b584de1fa3792bfd6d8f66e3ef975f4a
SERVICE_ROLE_KEY=eyJ...
```

**❌ MISSING in local .env:**
- `USE_HTTP_PLUGINS=true`
- `CRON_SCHEDULE=0 6 * * *`
- `PLUGIN_ARBETSFORMEDLINGEN_URL`
- `PLUGIN_EURES_URL`
- `PLUGIN_REMOTIVE_URL`
- `PLUGIN_REMOTEOK_URL`
- `PLUGIN_INDEED_CHROME_URL`
- `PLUGIN_JOOBLE_URL`

### 2. Easypanel Configuration Unknown

**Need to verify in Easypanel:**
- Is `USE_HTTP_PLUGINS=true` set in the main OpenJobs container?
- Is `CRON_SCHEDULE=0 6 * * *` set?
- Are all `PLUGIN_*_URL` variables configured?
- Are all plugin containers running?

### 3. Scheduler Behavior Without CRON_SCHEDULE

If `CRON_SCHEDULE` is not set, the scheduler falls back to **interval mode**:
- Default: Every 24 hours
- Starts immediately on container startup
- Then repeats every 24 hours from startup time

**This means:** If the container restarts at 2:00 PM, next sync is at 2:00 PM the next day (not 6:00 AM!)

---

## 🎯 What Should Happen

### Expected Flow (with CRON_SCHEDULE=0 6 * * *)

1. **Container starts** → Scheduler initializes
2. **Cron scheduler starts** → Waits for 6:00 AM
3. **At 6:00 AM** → Cron triggers
4. **Scheduler calls** → `RunManualSync()`
5. **For each plugin URL** → POST to `http://plugin-name:port/sync`
6. **Plugins sync** → Fetch jobs, store in database
7. **Logs show** → "✅ [Plugin Name] HTTP sync completed"

### Expected Logs

```
⏰ Starting job ingestion with cron schedule: 0 6 * * *
✅ Cron scheduler started
📅 Examples:
   '0 6 * * *'   - Every day at 6:00 AM
   '0 */6 * * *' - Every 6 hours
   '0 0 * * *'   - Every day at midnight

⏰ Running scheduled job sync at 2025-10-25 06:00:00
🔌 Using HTTP plugin containers (microservices mode)
🔧 Running manual job sync for all connectors...
✅ Arbetsförmedlingen HTTP sync completed
✅ EURES HTTP sync completed
✅ Remotive HTTP sync completed
✅ RemoteOK HTTP sync completed
✅ Indeed Chrome HTTP sync completed
✅ Jooble HTTP sync completed
✅ All scheduled syncs completed
```

---

## 🔧 Investigation Checklist

### In Easypanel Dashboard

- [ ] Check OpenJobs main container environment variables
- [ ] Verify `USE_HTTP_PLUGINS=true` is set
- [ ] Verify `CRON_SCHEDULE=0 6 * * *` is set
- [ ] Check all `PLUGIN_*_URL` variables are configured
- [ ] Verify all plugin containers are running
- [ ] Check container logs for cron trigger messages
- [ ] Check last sync time in database

### In OpenJobs_Web Dashboard

- [ ] Check "Last Sync" timestamps for each connector
- [ ] Verify sync logs show recent activity
- [ ] Check if sync times match 6:00 AM
- [ ] Look for error messages in sync logs

### Database Query

```sql
-- Check recent sync activity
SELECT 
    source,
    COUNT(*) as job_count,
    MAX(created_at) as last_sync
FROM job_posts
GROUP BY source
ORDER BY last_sync DESC;

-- Check sync logs (if table exists)
SELECT *
FROM sync_logs
ORDER BY created_at DESC
LIMIT 20;
```

---

## 💡 Recommended Actions

### 1. Verify Easypanel Configuration

**Access Easypanel** → OpenJobs container → Environment Variables

**Required variables:**
```bash
USE_HTTP_PLUGINS=true
CRON_SCHEDULE=0 6 * * *
PLUGIN_ARBETSFORMEDLINGEN_URL=http://arbetsformedlingen:8081
PLUGIN_EURES_URL=http://eures:8082
PLUGIN_REMOTIVE_URL=http://remotive:8083
PLUGIN_REMOTEOK_URL=http://remoteok:8084
PLUGIN_INDEED_CHROME_URL=http://indeed-chrome:8087
PLUGIN_JOOBLE_URL=http://jooble:8088
```

### 2. Check Container Logs

```bash
# In Easypanel, view logs for OpenJobs container
# Look for:
# - "Starting job ingestion with cron schedule: 0 6 * * *"
# - "Cron triggered at: [timestamp]"
# - "HTTP sync completed" messages
```

### 3. Manual Sync Test

```bash
# Test manual sync via API
curl -X POST https://app-openjobs.katsu6.easypanel.host/sync/manual

# Should trigger all plugins immediately
# Check logs for success/failure messages
```

### 4. Update Local .env for Testing

Add to `/Users/mafr/Code/github/openlazyjobs/OpenJobs/.env`:

```bash
# Microservices Mode
USE_HTTP_PLUGINS=true

# Cron Schedule (6 AM daily)
CRON_SCHEDULE=0 6 * * *

# Plugin URLs (for local Docker Compose testing)
USE_LOCALHOST_DEFAULTS=true
PLUGIN_ARBETSFORMEDLINGEN_URL=http://localhost:8081
PLUGIN_EURES_URL=http://localhost:8082
PLUGIN_REMOTIVE_URL=http://localhost:8083
PLUGIN_REMOTEOK_URL=http://localhost:8084
PLUGIN_INDEED_CHROME_URL=http://localhost:8087
PLUGIN_JOOBLE_URL=http://localhost:8088
```

---

## 🚨 Common Problems

### Problem 1: Plugins Running on Their Own Schedule

**Symptom:** Each plugin has its own `CRON_SCHEDULE` in their container

**Solution:** Remove `CRON_SCHEDULE` from plugin containers - only the main container should have it

### Problem 2: Interval Mode Instead of Cron

**Symptom:** Syncs happen 24 hours after container restart, not at 6 AM

**Solution:** Ensure `CRON_SCHEDULE=0 6 * * *` is set in main container

### Problem 3: Plugins Not Reachable

**Symptom:** "HTTP sync failed" errors in logs

**Solution:** 
- Verify plugin containers are running
- Check internal network names match `PLUGIN_*_URL` variables
- Test connectivity: `curl http://plugin-name:port/health`

### Problem 4: USE_HTTP_PLUGINS Not Set

**Symptom:** Local connectors run instead of calling HTTP plugins

**Solution:** Set `USE_HTTP_PLUGINS=true` in main container

---

## 📊 Next Steps

1. **Check Easypanel** - Verify all environment variables
2. **Review logs** - Look for cron trigger messages
3. **Check dashboard** - See last sync times
4. **Test manual sync** - Verify plugins respond
5. **Fix configuration** - Add missing env vars if needed
6. **Monitor** - Watch for next 6 AM sync

---

**Status:** 🔍 Investigation in progress  
**Action Required:** Check Easypanel environment variables and container logs
