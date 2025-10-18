# Remotive SSL Certificate Issue

## Problem

Remotive.io API is experiencing SSL certificate issues (Cloudflare Error 526):

```
Error 526: Invalid SSL certificate
The origin web server does not have a valid SSL certificate.
```

## Impact

- Remotive connector cannot fetch jobs
- This is a **Remotive server issue**, not an OpenJobs problem
- Other connectors (Arbetsförmedlingen, EURES, RemoteOK) continue working normally

## Solution Implemented

Added graceful error handling to the Remotive connector:

### Before (Failed Entire Sync)
```go
resp, err := client.Do(req)
if err != nil {
    return nil, fmt.Errorf("failed to fetch jobs: %w", err)
}
```

**Result:** Entire sync failed, no jobs from any connector

### After (Graceful Degradation)
```go
resp, err := client.Do(req)
if err != nil {
    fmt.Printf("⚠️  Remotive API unavailable (SSL/connection error)\n")
    fmt.Println("   Skipping Remotive sync - will retry next cycle")
    return []models.JobPost{}, nil // Return empty, don't fail
}

// Also handle Cloudflare 526 errors
if resp.StatusCode == 526 || strings.Contains(body, "Invalid SSL certificate") {
    fmt.Printf("⚠️  Remotive API has SSL certificate issues (Error 526)\n")
    return []models.JobPost{}, nil
}
```

**Result:** Remotive skipped, other connectors continue normally

## Current Behavior

When Remotive API is down:

1. ✅ Arbetsförmedlingen syncs normally
2. ✅ EURES syncs normally  
3. ⚠️  Remotive skips (logs warning)
4. ✅ RemoteOK syncs normally

**Dashboard shows:**
- Remotive: 0 fetched, 0 inserted (skipped due to SSL error)
- Other connectors: Normal operation

## Monitoring

### Check Remotive Status

```bash
# Test Remotive API directly
curl -I https://remotive.io/api/remote-jobs

# Expected when working:
HTTP/2 200

# Current error:
HTTP/2 526
```

### Check OpenJobs Logs

```bash
# Look for Remotive warnings
docker logs openjobs-remotive 2>&1 | grep "⚠️"

# Expected output when SSL issue:
⚠️  Remotive API unavailable (SSL/connection error)
   Skipping Remotive sync - will retry next cycle
```

## When Will It Be Fixed?

This is a **Remotive infrastructure issue**. They need to:
1. Fix their SSL certificate
2. Update Cloudflare configuration
3. Ensure certificate is valid and trusted

**ETA:** Unknown - depends on Remotive team

## Workarounds

### Option 1: Wait (Recommended)
- Remotive will fix their SSL certificate
- OpenJobs will automatically resume syncing
- No action needed on your part

### Option 2: Disable Remotive Temporarily
If you don't want the warning logs:

```bash
# In Easypanel, stop the Remotive plugin container
# Or set environment variable:
DISABLE_REMOTIVE=true
```

### Option 3: Alternative API
Remotive might have an alternative endpoint:

```go
// Try different base URL (if available)
baseURL: "https://api.remotive.io"  // Instead of remotive.io/api
```

## Testing When Fixed

Once Remotive fixes their SSL:

```bash
# Test manually
curl https://remotive.io/api/remote-jobs?limit=1

# If successful, trigger sync
curl -X POST http://your-openjobs/sync/manual
```

## Related Issues

- Cloudflare Error 526: https://developers.cloudflare.com/support/troubleshooting/cloudflare-errors/troubleshooting-cloudflare-5xx-errors/#error-526-invalid-ssl-certificate
- Remotive Status: Check their status page or Twitter

## Files Modified

- `connectors/remotive/connector.go` - Added graceful error handling

## Summary

✅ **OpenJobs is working correctly**
❌ **Remotive API has SSL issues** (external problem)
✅ **Other connectors unaffected**
⏳ **Will auto-resume when Remotive fixes their certificate**

No action required from you - the system will automatically retry and resume when Remotive is back online.
