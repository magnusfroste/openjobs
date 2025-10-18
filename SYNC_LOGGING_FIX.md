# Sync Logging Fix for RemoteOK and Remotive

## Problem
RemoteOK and Remotive plugins were syncing jobs but **not creating sync log entries** in the database. This caused them to be missing from the dashboard sync history.

## Root Cause
The `SyncJobs()` method in both connectors was missing calls to `store.LogSync()`.

## Solution
Added sync logging to both connectors, matching the pattern used in Arbetsförmedlingen and EURES.

### Changes Made

**Files Modified:**
1. `/Users/mafr/Code/OpenJobs/connectors/remoteok/connector.go`
2. `/Users/mafr/Code/OpenJobs/connectors/remotive/connector.go`

**What Was Added:**
1. Track `startTime` at beginning of sync
2. Log failed syncs with error status
3. Track `duplicates` count
4. Log successful syncs with full statistics

### Before
```go
func (rc *RemoteOKConnector) SyncJobs() error {
    jobs, err := rc.FetchJobs()
    if err != nil {
        return err  // No logging!
    }
    
    stored := 0
    for _, job := range jobs {
        // ... store jobs ...
    }
    
    return nil  // No logging!
}
```

### After
```go
func (rc *RemoteOKConnector) SyncJobs() error {
    startTime := time.Now()  // Track start
    
    jobs, err := rc.FetchJobs()
    if err != nil {
        // Log failed sync
        rc.store.LogSync(&models.SyncLog{
            ConnectorName: rc.GetID(),
            StartedAt:     startTime,
            CompletedAt:   time.Now(),
            JobsFetched:   0,
            JobsInserted:  0,
            JobsDuplicates: 0,
            Status:        "failed",
        })
        return err
    }
    
    stored := 0
    duplicates := 0  // Track duplicates
    for _, job := range jobs {
        if existing != nil {
            duplicates++  // Count duplicates
            continue
        }
        // ... store jobs ...
    }
    
    // Log successful sync
    rc.store.LogSync(&models.SyncLog{
        ConnectorName:  rc.GetID(),
        StartedAt:      startTime,
        CompletedAt:    time.Now(),
        JobsFetched:    len(jobs),
        JobsInserted:   stored,
        JobsDuplicates: duplicates,
        Status:         "success",
    })
    
    return nil
}
```

## Expected Result

After redeploying the plugin containers, the dashboard sync history will show all 4 plugins:

```
Plugin                Time          Fetched  Inserted  Duplicates  Efficiency
Arbetsförmedlingen   03:05 18/10   20       0         20          0%
EURES                03:05 18/10   10       0         10          0%
RemoteOK             03:05 18/10   50       0         50          0%
Remotive             03:05 18/10   0        0         0           0%
```

## Deployment

### For Plugin Containers (Easypanel)

Rebuild and redeploy both plugin containers:

1. **RemoteOK Plugin:**
   - Service: `app_openjobs-remoteok`
   - Rebuild from latest code
   - Restart container

2. **Remotive Plugin:**
   - Service: `app_openjobs-remotive`
   - Rebuild from latest code
   - Restart container

### For Local Testing

```bash
# Stop current server
pkill -f "openjobs"

# Rebuild and run
cd /Users/mafr/Code/OpenJobs
./run-local.sh
```

Then trigger a manual sync:
```bash
curl -X POST http://localhost:9090/sync/manual
```

Check sync logs:
```bash
curl -s http://localhost:9090/sync/logs | jq '.data[0:4] | .[] | {connector: .connector_name, fetched: .jobs_fetched}'
```

Should now show all 4 connectors!

## Testing

### Verify Logging Works

```bash
# Trigger sync on RemoteOK plugin directly
curl -X POST https://app-openjobs-remoteok.katsu6.easypanel.host/sync

# Check if log entry was created
curl -s https://app-openjobs.katsu6.easypanel.host/sync/logs | \
  jq '.data[] | select(.connector_name == "remoteok") | {time: .started_at, fetched: .jobs_fetched}'
```

### Dashboard Check

After deployment, the dashboard should show:
- ✅ All 4 plugins in sync history
- ✅ Real fetch/insert/duplicate counts
- ✅ Accurate timestamps
- ✅ Efficiency percentages

## Notes

- This fix only affects the **plugin containers**, not the core container
- The core container doesn't need to be rebuilt for this fix
- Sync logs are stored in the `sync_logs` table in Supabase
- Each sync creates one log entry per connector
