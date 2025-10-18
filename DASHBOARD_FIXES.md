# OpenJobs Dashboard Fixes

## Issues Fixed

### 1. ✅ Duplicate Sync Logs
**Problem:** Connectors running twice - once in core container, once in plugin containers

**Solution:** Added `USE_HTTP_PLUGINS` environment variable
- When `USE_HTTP_PLUGINS=true`: Core scheduler calls plugin containers via HTTP
- When `USE_HTTP_PLUGINS=false`: Core runs local connectors (monolith mode)

**Files Changed:**
- `internal/scheduler/scheduler.go` - Added mode detection in `runSync()`

### 2. ✅ Incorrect Job Count
**Problem:** Dashboard showed hardcoded mock data (48 jobs) instead of real count

**Solution:** Added database query methods
- `GetTotalJobCount()` - Returns actual job count from database
- `GetRemoteJobCount()` - Returns remote job count
- Analytics endpoint now queries real data

**Files Changed:**
- `pkg/storage/job.go` - Added count methods
- `internal/api/handlers.go` - AnalyticsHandler now uses real data

### 3. ✅ Wrong "Last Sync" Timestamp
**Problem:** "Last Sync" updated on every page refresh (showed current time)

**Solution:** Fetch actual last sync time from sync_logs table
- Dashboard now queries `GetRecentSyncLogs(1)` 
- Shows actual sync time, not page load time
- Format: HH:MM DD/MM (e.g., "22:30 18/10")

**Files Changed:**
- `internal/api/handlers.go` - Fixed dashboard JavaScript and analytics endpoint

### 4. ✅ Time Format Changed
**Problem:** Relative time ("3h ago") hard to compare

**Solution:** Changed to absolute time format
- Old: "3h ago", "2d ago"
- New: "22:30 18/10" (time + date)

**Files Changed:**
- `internal/api/handlers.go` - Updated `getTimeAgo()` JavaScript function

## Deployment Instructions

### Environment Variables for Core Container

Add to your OpenJobs core container in Easypanel:

```bash
# Microservices mode - call plugin containers via HTTP
USE_HTTP_PLUGINS=true

# Plugin container URLs (use internal Docker network names)
PLUGIN_ARBETSFORMEDLINGEN_URL=http://plugin-arbetsformedlingen:8081
PLUGIN_EURES_URL=http://plugin-eures:8082
PLUGIN_REMOTIVE_URL=http://plugin-remotive:8083
PLUGIN_REMOTEOK_URL=http://plugin-remoteok:8084
```

### Rebuild and Deploy

```bash
cd /Users/mafr/Code/OpenJobs
go build -o openjobs cmd/openjobs/main.go
# Deploy to Easypanel
```

## Expected Results

### Dashboard Should Show:
- ✅ **Total Jobs**: Real count from database (e.g., 145)
- ✅ **Last Sync**: Actual sync time (e.g., "01:57 18/10")
- ✅ **Sync History**: Each connector appears ONCE per sync
- ✅ **All 4 Plugins**: Arbetsförmedlingen, EURES, Remotive, RemoteOK

### Sync History Table:
```
Plugin                Time          Fetched  Inserted  Duplicates  Efficiency
Arbetsförmedlingen   01:57 18/10   20       5         15          25%
EURES                01:57 18/10   3        0         3           0%
Remotive             01:57 18/10   50       30        20          60%
RemoteOK             01:57 18/10   72       45        27          62%
```

## Architecture

### Microservices Mode (Current)
```
Core Container (port 8080)
  ├─ Scheduler (every 6h)
  │   ├─ POST http://plugin-arbetsformedlingen:8081/sync
  │   ├─ POST http://plugin-eures:8082/sync
  │   ├─ POST http://plugin-remotive:8083/sync
  │   └─ POST http://plugin-remoteok:8084/sync
  ├─ Dashboard
  └─ API

Plugin Containers
  ├─ plugin-arbetsformedlingen:8081 (waits for POST /sync)
  ├─ plugin-eures:8082 (waits for POST /sync)
  ├─ plugin-remotive:8083 (waits for POST /sync)
  └─ plugin-remoteok:8084 (waits for POST /sync)
```

### Monolith Mode (Alternative)
```
Core Container (port 8080)
  ├─ Scheduler (every 6h)
  │   ├─ Run Arbetsförmedlingen locally
  │   ├─ Run EURES locally
  │   ├─ Run Remotive locally
  │   └─ Run RemoteOK locally
  ├─ Dashboard
  └─ API
```

Set `USE_HTTP_PLUGINS=false` for monolith mode.

## Troubleshooting

### Still Seeing Duplicates?
- Check only ONE core container is running
- Verify `USE_HTTP_PLUGINS=true` is set
- Check plugin URLs are correct

### Plugin Not Showing in Logs?
- Check plugin container is running
- Check plugin health: `curl http://plugin-name:8081/health`
- Check core can reach plugin: `curl -X POST http://plugin-name:8081/sync`

### Wrong Job Count?
- Verify Supabase connection
- Check `SUPABASE_URL` and `SUPABASE_ANON_KEY` are set
- Test: `curl https://your-supabase.supabase.co/rest/v1/job_posts?select=count`

### "Last Sync" Still Wrong?
- Check sync_logs table has data
- Verify GetRecentSyncLogs() returns results
- Check browser console for JavaScript errors
