# Incident Report: Missing Jobs from 2025-11-28 04:00 Sync

**Date:** 2025-11-28  
**Time:** 04:00 UTC  
**Status:** ✅ RESOLVED  

## Summary

Jobs synced during the 2025-11-28 04:00 cron run appeared to be missing from OpenJobs, but were actually inserted with incorrect timestamps (`created_at = 0001-01-01`), making them invisible to date-based queries.

## Timeline

- **04:00** - Cron triggered scheduled sync
- **04:00** - Arbetsförmedlingen plugin fetched 37 jobs, inserted 12 new jobs
- **04:00** - Logs showed successful inserts: `✅ Job created successfully (status: 201)`
- **12:49** - User reported jobs missing from database
- **12:49-13:30** - Investigation and fix deployed

## Root Cause

The `JobPost` struct in `pkg/models/job.go` had `CreatedAt` and `UpdatedAt` fields without `omitempty` JSON tags:

```go
// BEFORE (BROKEN)
CreatedAt time.Time `json:"created_at" db:"created_at"`
UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
```

**Impact:**
1. Connectors don't explicitly set these fields when creating jobs
2. Fields default to Go's zero time value (`0001-01-01T00:00:00Z`)
3. Zero time is marshaled to JSON and sent to Supabase
4. Overrides database's `DEFAULT NOW()` constraint
5. Jobs stored with `created_at = 0001-01-01`
6. Invisible to queries filtering by `created_at >= <date>`

## Affected Data

- **Total jobs affected:** 20 (out of 813 total)
- **Sources affected:** All connectors
  - Arbetsförmedlingen: 12 jobs
  - EURES/Adzuna: 3 jobs
  - Remotive: 2 jobs
  - RemoteOK: 3 jobs

### Example Affected Jobs
```
af-30315109 - Back-end developer
af-30314831 - Systems Programmer
af-30314820 - Gameplay Programmer
adzuna-5515930537
remotive-2080593
remoteok-1129056
```

## Fix Applied

### 1. Code Fix (✅ COMPLETED)
**File:** `pkg/models/job.go`  
**Change:** Added `omitempty` to JSON tags

```go
// AFTER (FIXED)
CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
```

**Effect:** Zero time values now excluded from JSON payload, allowing database defaults to work.

### 2. Data Migration (✅ COMPLETED)
**File:** `migrations/007_fix_zero_timestamps.sql`  
**Script:** `scripts/apply_migration_007.sh`

Updated 20 jobs with zero timestamps:
- Set `created_at` = `posted_date` (more accurate than NOW())
- Set `updated_at` = `posted_date`

**Result:**
```bash
Before: 20 jobs with created_at < 2000-01-01
After:  0 jobs with created_at < 2000-01-01
```

### 3. Deployment (⏳ PENDING)
Plugins need to be rebuilt and redeployed to use the fixed code:

```bash
# Rebuild all plugin containers
docker-compose -f docker-compose.plugins.yml build

# Restart all services
docker-compose -f docker-compose.plugins.yml up -d
```

## Verification

### Before Fix
```bash
curl "https://cmpnqpdxhmecptcbffmw.supabase.co/rest/v1/job_posts?select=id,created_at&id=eq.af-30315109" \
  -H "apikey: YOUR_KEY"
# Result: {"id":"af-30315109","created_at":"0001-01-01T00:00:00+00:00"}
```

### After Fix
```bash
curl "https://cmpnqpdxhmecptcbffmw.supabase.co/rest/v1/job_posts?select=id,created_at&id=eq.af-30315109" \
  -H "apikey: YOUR_KEY"
# Result: {"id":"af-30315109","created_at":"2025-11-27T17:45:12+00:00"}
```

### Query Results
```bash
# Jobs from Nov 27-28 with is_active=true
curl "https://cmpnqpdxhmecptcbffmw.supabase.co/rest/v1/job_posts?select=count&created_at=gte.2025-11-27T00:00:00&is_active=eq.true" \
  -H "apikey: YOUR_KEY" -H "Prefer: count=exact"
# Result: 64 jobs (includes the 20 fixed jobs)
```

## Impact Assessment

### User Impact
- **LazyJobs connector:** Could not fetch new jobs from OpenJobs API
- **API consumers:** Date-filtered queries returned incomplete results
- **Dashboard:** Job counts and analytics were incorrect

### Duration
- **Incident duration:** ~8 hours (04:00 - 12:49)
- **Resolution time:** ~45 minutes (12:49 - 13:30)

### Data Integrity
- ✅ No data loss - all jobs were stored
- ✅ All job data intact (title, company, description, etc.)
- ✅ Only timestamps affected
- ✅ Fixed timestamps use `posted_date` (accurate)

## Lessons Learned

### What Went Well
1. Logs clearly showed successful inserts
2. Quick identification of root cause
3. Clean fix with minimal code change
4. Successful data recovery using `posted_date`

### What Could Be Improved
1. **Testing:** Add integration tests for timestamp handling
2. **Monitoring:** Alert on jobs with invalid timestamps
3. **Validation:** Add database constraint to reject dates before 2000
4. **Documentation:** Document timestamp handling in connector guide

## Action Items

- [x] Fix code (add `omitempty` tags)
- [x] Create migration script
- [x] Apply migration to fix existing data
- [x] Verify fix
- [x] Commit and push changes
- [ ] Rebuild and redeploy plugin containers
- [ ] Add integration test for timestamp handling
- [ ] Add database constraint: `CHECK (created_at >= '2000-01-01')`
- [ ] Update connector development guide

## Related Files

- `pkg/models/job.go` - Fixed struct tags
- `migrations/007_fix_zero_timestamps.sql` - SQL migration
- `scripts/apply_migration_007.sh` - Migration script
- `BUGFIX_ZERO_TIMESTAMPS.md` - Detailed technical documentation

## Commits

- `858eefa` - Fix: Add omitempty to CreatedAt/UpdatedAt
- `b38e161` - Add migration script to fix zero timestamps via REST API

## References

- GitHub Issue: N/A (direct user report)
- Supabase Project: cmpnqpdxhmecptcbffmw
- Affected API: `/rest/v1/job_posts`
