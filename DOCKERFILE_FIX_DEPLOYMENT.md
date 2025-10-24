# Dockerfile Fix - Remove Hardcoded Cron Schedule

**Date:** October 25, 2025  
**Issue:** Plugins syncing every 6 hours instead of once daily at 6 AM  
**Root Cause:** Dockerfile had hardcoded supercronic with `0 */6 * * *` schedule

---

## 🔍 Problem Summary

### What Was Happening

The Dockerfile had **two scheduling systems running simultaneously**:

1. **Supercronic** (external cron) - Hardcoded to `0 */6 * * *` (every 6 hours)
   - Line 32: `RUN echo "0 */6 * * * curl -X POST http://localhost:8080/sync/manual..."`
   - This was **actually triggering** the syncs

2. **Go Internal Scheduler** - Reads `CRON_SCHEDULE` env var (set to `0 6 * * *`)
   - This was **being ignored** because supercronic triggered first

**Result:** Plugins synced at 08:00, 14:00, 20:00, 02:00 (every 6 hours) instead of once at 06:00.

---

## ✅ Solution

**Removed supercronic entirely** - The Go application has its own robust internal scheduler that:
- Reads `CRON_SCHEDULE` environment variable
- Supports standard cron syntax
- Calls all plugins via HTTP at the specified time
- Logs all sync activity

### Changes Made to Dockerfile

1. **Removed supercronic installation** (lines 21-24)
2. **Removed hardcoded crontab** (line 32)
3. **Removed startup script** (lines 35-47)
4. **Simplified CMD** to run Go app directly

### New Dockerfile Structure

```dockerfile
FROM golang:1.23-alpine AS builder
# ... build steps ...

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/openjobs .

# Scheduling handled by Go app's internal scheduler
# Set CRON_SCHEDULE environment variable to control sync frequency

EXPOSE 8080
CMD ["./openjobs"]
```

---

## 🚀 Deployment Steps

### 1. Commit Changes

```bash
cd /Users/mafr/Code/github/openlazyjobs/OpenJobs
git add Dockerfile
git commit -m "Fix: Remove hardcoded cron, use internal Go scheduler"
git push
```

### 2. Rebuild Container in Easypanel

**Option A: Trigger Rebuild**
- Go to Easypanel → OpenJobs app
- Click "Rebuild" or "Redeploy"
- Wait for build to complete

**Option B: Manual Docker Build**
```bash
docker build -t openjobs:latest .
docker push your-registry/openjobs:latest
# Update Easypanel to use new image
```

### 3. Verify Environment Variables

Ensure these are still set in Easypanel:
```bash
USE_HTTP_PLUGINS=true
CRON_SCHEDULE=0 6 * * *
PLUGIN_ARBETSFORMEDLINGEN_URL=http://app_openjobs-arbfrm:8081
PLUGIN_EURES_URL=http://app_openjobs-eures:8082
PLUGIN_REMOTIVE_URL=http://app_openjobs-remotive:8083
PLUGIN_REMOTEOK_URL=http://app_openjobs-remoteok:8084
PLUGIN_INDEED_CHROME_URL=http://app_openjobs-indeed-chrome:8087
PLUGIN_JOOBLE_URL=http://app_openjobs-jooble:8088
```

### 4. Check Logs After Deployment

Look for this message in container logs:
```
⏰ Starting job ingestion with cron schedule: 0 6 * * *
✅ Cron scheduler started
📅 Examples:
   '0 6 * * *'   - Every day at 6:00 AM
   '0 */6 * * *' - Every 6 hours
   '0 0 * * *'   - Every day at midnight
```

**NOT this:**
```
⏰ Starting cron scheduler (job sync every 6 hours)...
```

---

## 🎯 Expected Behavior After Fix

### Sync Schedule
- **Once per day at 6:00 AM** (container timezone)
- No more 14:00 or 20:00 syncs

### Logs at 6:00 AM
```
⏰ Cron triggered at: 2025-10-25 06:00:00
🔧 Running manual job sync for all connectors...
🔌 Using HTTP plugin containers (microservices mode)
✅ Arbetsförmedlingen HTTP sync completed
✅ EURES HTTP sync completed
✅ Remotive HTTP sync completed
✅ RemoteOK HTTP sync completed
✅ Indeed Chrome HTTP sync completed
✅ Jooble HTTP sync completed
✅ All scheduled syncs completed
```

### OpenJobs_Web Dashboard
Next sync times should show:
- Oct 25, 2025, 06:00 (next morning)
- Oct 26, 2025, 06:00 (day after)
- Oct 27, 2025, 06:00 (etc.)

---

## 🔧 Testing

### Manual Sync Test (Before 6 AM)
```bash
curl -X POST https://app-openjobs.katsu6.easypanel.host/sync/manual
```

This should trigger all plugins immediately and you'll see the sync in the dashboard.

### Wait for Automatic Sync
- Wait until 6:00 AM the next day
- Check OpenJobs_Web dashboard
- Should see all plugins synced at 06:00

---

## 🌍 Timezone Considerations

### Current Setup
Container likely runs in **UTC timezone**.

**6:00 AM UTC = 8:00 AM CEST (Sweden summer time)**

### If You Want 6 AM Sweden Time

Add to Easypanel environment variables:
```bash
TZ=Europe/Stockholm
CRON_SCHEDULE=0 6 * * *  # Now 6 AM Stockholm time
```

Or adjust cron to UTC equivalent:
```bash
CRON_SCHEDULE=0 4 * * *  # 4 AM UTC = 6 AM CEST
```

---

## 📊 Monitoring

### Check Sync History
Go to OpenJobs_Web dashboard → Recent Sync History

**Before fix:**
- Syncs at 08:00, 14:00, 20:00, 02:00

**After fix:**
- Syncs only at 06:00 (or 08:00 if UTC)

### Check Container Logs
```bash
# In Easypanel, view logs
# Look for cron trigger messages at 6 AM only
```

---

## 🎉 Benefits

### ✅ Cleaner Architecture
- Single scheduling system (Go internal scheduler)
- No external dependencies (supercronic removed)
- Simpler Dockerfile

### ✅ Configurable
- Change schedule via `CRON_SCHEDULE` env var
- No need to rebuild container
- Supports any cron syntax

### ✅ Better Logging
- Go scheduler logs all sync activity
- Easier to debug
- Clear error messages

### ✅ Consistent
- All plugins sync together at same time
- No race conditions
- Predictable behavior

---

## 🔄 Rollback (If Needed)

If something goes wrong, you can rollback to the old Dockerfile:

```bash
git revert HEAD
git push
# Rebuild in Easypanel
```

But the new approach is **cleaner and more maintainable**.

---

## 📝 Summary

**What we fixed:**
- ❌ Removed hardcoded `0 */6 * * *` cron from Dockerfile
- ❌ Removed supercronic dependency
- ❌ Removed complex startup script
- ✅ Now uses Go's internal scheduler
- ✅ Respects `CRON_SCHEDULE` environment variable
- ✅ Simpler, cleaner, more maintainable

**Result:**
- Plugins will sync **once per day at 6:00 AM** as intended
- No more unexpected 14:00 and 20:00 syncs
- Schedule is now fully configurable via environment variables

---

**Status:** ✅ Ready to deploy  
**Action Required:** Commit, push, and rebuild OpenJobs container in Easypanel
