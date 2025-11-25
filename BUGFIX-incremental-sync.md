# Bug Fix: LazyJobs OpenJobs Connector Incremental Sync

## Issue
The LazyJobs OpenJobs connector was not finding new jobs during incremental sync, even though jobs existed in the OpenJobs database.

### Symptoms
```
📅 Last sync: 2025-11-24T08:50:23.620Z
🌐 Fetching jobs created after 2025-11-24T08:50:23.620Z...
✅ Fetched 0 jobs from OpenJobs
```

## Root Cause

The `JobPost` model in OpenJobs was **missing `created_at` and `updated_at` fields** in the JSON response, even though:

1. ✅ The database columns exist (`created_at`, `updated_at`)
2. ✅ The database has proper indexes
3. ✅ The API filter works correctly (`created_at=gte.{timestamp}`)
4. ❌ **The JSON response didn't include these timestamps**

### Before Fix

```go
type JobPost struct {
    ID              string                 `json:"id" db:"id"`
    Title           string                 `json:"title" db:"title"`
    // ... other fields ...
    Source          string                 `json:"source" db:"source"`
    Fields          map[string]interface{} `json:"fields,omitempty" db:"fields"`
    // ❌ Missing: CreatedAt and UpdatedAt
}
```

API Response:
```json
{
  "id": "af-30274728",
  "title": "VHDL & FPGA Developer",
  "posted_date": "2025-11-17T23:26:25Z",
  // ❌ No created_at field!
}
```

### After Fix

```go
type JobPost struct {
    ID              string                 `json:"id" db:"id"`
    Title           string                 `json:"title" db:"title"`
    // ... other fields ...
    Source          string                 `json:"source" db:"source"`
    Fields          map[string]interface{} `json:"fields,omitempty" db:"fields"`
    CreatedAt       time.Time              `json:"created_at" db:"created_at"`   // ✅ Added
    UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`   // ✅ Added
}
```

Expected API Response:
```json
{
  "id": "af-30274728",
  "title": "VHDL & FPGA Developer",
  "posted_date": "2025-11-17T23:26:25Z",
  "created_at": "2025-11-17T23:30:00Z",  // ✅ Now visible!
  "updated_at": "2025-11-17T23:30:00Z"
}
```

## Why This Matters

### Incremental Sync Flow

1. **LazyJobs connector** reads last sync time from database
2. Calls OpenJobs API: `GET /jobs?created_after=2025-11-24T08:50:23.620Z`
3. **OpenJobs API** filters database: `WHERE created_at >= '2025-11-24T08:50:23.620Z'`
4. Returns matching jobs in JSON
5. **LazyJobs** processes jobs and saves new sync time

### The Problem

Without `created_at` in the JSON response:
- ❌ LazyJobs cannot verify which jobs are actually new
- ❌ Cannot debug sync issues (no visibility into when jobs were added)
- ❌ Cannot implement client-side filtering or validation
- ❌ Incremental sync relies entirely on server-side filtering (black box)

### The Solution

With `created_at` in the JSON response:
- ✅ LazyJobs can see exactly when each job was added to OpenJobs
- ✅ Can implement client-side validation and filtering
- ✅ Better debugging and monitoring
- ✅ Transparent incremental sync

## Files Changed

- `pkg/models/job.go` - Added `CreatedAt` and `UpdatedAt` fields to `JobPost` struct

## Testing

### Before Deployment
```bash
# Build test
go build -o /tmp/openjobs-test ./cmd/openjobs/main.go
```

### After Deployment
```bash
# Verify timestamps are returned
curl -s "https://api.openjobs.ink/jobs?limit=1" | jq '.data[0] | {id, title, created_at, updated_at}'

# Expected output:
{
  "id": "af-30274728",
  "title": "VHDL & FPGA Developer",
  "created_at": "2025-11-17T23:30:00Z",
  "updated_at": "2025-11-17T23:30:00Z"
}
```

## Deployment Steps

1. ✅ Code committed and pushed to main
2. ⏳ Rebuild OpenJobs Docker image
3. ⏳ Deploy to production (Easypanel)
4. ⏳ Verify API response includes timestamps
5. ⏳ Monitor LazyJobs connector next sync

## Impact

- **Breaking Change**: No (additive change only)
- **API Version**: No version bump needed (backward compatible)
- **Database Migration**: None required (columns already exist)
- **Downtime**: None (rolling deployment)

## Related Files

- `/OpenJobs/pkg/models/job.go` - Model definition
- `/OpenJobs/internal/api/handlers.go` - API handlers
- `/OpenJobs/pkg/storage/job.go` - Database queries
- `/OpenJobs/migrations/001_create_job_posts.sql` - Database schema
- `/lazyjobs-lovable/connectors/openjobs/fetch-jobs.js` - LazyJobs connector

## Commit

```
commit f6835a8
Author: Magnus Froste
Date: 2025-11-26

fix: Add created_at and updated_at to JobPost API response

- Added CreatedAt and UpdatedAt fields to JobPost model
- These timestamps are critical for incremental sync in LazyJobs connector
- Without these fields, the connector cannot determine when jobs were added to OpenJobs
- Fixes issue where created_after filter worked but timestamps weren't visible to clients
```
