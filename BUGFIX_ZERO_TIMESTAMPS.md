# Bug Fix: Zero Timestamps (0001-01-01) in created_at/updated_at

## Problem

Jobs inserted on 2025-11-28 04:00 (and potentially earlier) had `created_at` and `updated_at` set to `0001-01-01T00:00:00Z` instead of the current timestamp.

### Root Cause

The `JobPost` struct in `pkg/models/job.go` had `CreatedAt` and `UpdatedAt` fields without `omitempty` JSON tags:

```go
CreatedAt time.Time `json:"created_at" db:"created_at"`
UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
```

When connectors create jobs, they don't explicitly set these fields, so they default to Go's zero time value. This zero time was marshaled as `"0001-01-01T00:00:00Z"` and sent to Supabase, **overriding** the database's `DEFAULT NOW()` constraint.

### Impact

- **20 jobs** had zero timestamps (out of 813 total)
- These jobs were invisible to queries filtering by `created_at >= <date>`
- LazyJobs connector couldn't find these jobs
- API endpoints filtering by date returned incomplete results

## Solution

### 1. Code Fix (✅ COMPLETED)

Added `omitempty` to JSON tags in `pkg/models/job.go`:

```go
CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
```

Now when these fields are zero, they're excluded from the JSON payload, allowing the database's `DEFAULT NOW()` to work correctly.

### 2. Database Migration (⏳ PENDING)

Created `migrations/007_fix_zero_timestamps.sql` to fix existing jobs:

```sql
UPDATE job_posts 
SET 
    created_at = COALESCE(
        CASE WHEN created_at < '2000-01-01' THEN posted_date ELSE created_at END,
        NOW()
    ),
    updated_at = COALESCE(
        CASE WHEN updated_at < '2000-01-01' THEN posted_date ELSE updated_at END,
        NOW()
    )
WHERE created_at < '2000-01-01' OR updated_at < '2000-01-01';
```

This uses `posted_date` as a fallback for `created_at` since it's more accurate than `NOW()`.

## Deployment Steps

1. **Deploy code fix** (rebuild and restart all plugin containers)
2. **Run migration** (apply `007_fix_zero_timestamps.sql` to Supabase)
3. **Verify** (check that all jobs have valid timestamps)

## Verification

### Before Fix
```bash
curl "https://cmpnqpdxhmecptcbffmw.supabase.co/rest/v1/job_posts?select=count&created_at=lt.2000-01-01" \
  -H "apikey: YOUR_KEY" -H "Prefer: count=exact"
# Result: [{"count":20}]
```

### After Fix
```bash
curl "https://cmpnqpdxhmecptcbffmw.supabase.co/rest/v1/job_posts?select=count&created_at=lt.2000-01-01" \
  -H "apikey: YOUR_KEY" -H "Prefer: count=exact"
# Expected: [{"count":0}]
```

## Affected Connectors

All connectors were affected since none explicitly set `CreatedAt`:
- ✅ arbetsformedlingen
- ✅ eures
- ✅ remotive
- ✅ remoteok
- ✅ indeed
- ✅ indeed-scraper
- ✅ indeed-chrome
- ✅ jooble
- ✅ offentligajobb

## Logs from Incident

From 2025-11-28 04:00:00 sync:
```
✅ Arbetsförmedlingen sync complete! Fetched: 37, Inserted: 12, Duplicates: 25
```

The 12 jobs were successfully inserted but with zero timestamps:
- af-30315109 (Back-end developer)
- af-30314831 (Systems Programmer)
- af-30314820 (Gameplay Programmer)
- ... (9 more)

## Prevention

The `omitempty` tag ensures this won't happen again. Future jobs will:
1. Not include `created_at`/`updated_at` in JSON payload
2. Rely on database `DEFAULT NOW()` constraint
3. Get correct timestamps automatically

## Related Files

- `pkg/models/job.go` - Fixed struct tags
- `migrations/007_fix_zero_timestamps.sql` - Database fix
- `connectors/*/connector.go` - All connectors (no changes needed)
