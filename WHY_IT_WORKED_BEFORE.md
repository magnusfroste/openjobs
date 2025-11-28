# Why Did This Work for Weeks Before?

## Timeline of Events

### ✅ **Nov 13 - Nov 25: Everything Working**
Jobs were inserted with correct timestamps:
```json
{
  "id": "remotive-2074545",
  "created_at": "2025-11-13T06:00:01.098417+00:00",  // ✅ Correct
  "posted_date": "2025-11-13T06:00:01.001503+00:00"
}
```

**Why it worked:**
- `CreatedAt` and `UpdatedAt` fields **didn't exist** in the JobPost struct
- Connectors couldn't set them (they weren't there)
- Database `DEFAULT NOW()` constraint worked perfectly
- No JSON fields to override the defaults

### 🔧 **Nov 26, 00:36 (Commit f6835a8): The Breaking Change**

**What happened:**
```diff
// pkg/models/job.go
  Source          string                 `json:"source" db:"source"`
  Fields          map[string]interface{} `json:"fields,omitempty" db:"fields"`
+ CreatedAt       time.Time              `json:"created_at" db:"created_at"`   // ⚠️ NO omitempty!
+ UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`   // ⚠️ NO omitempty!
}
```

**Why the change was made:**
- LazyJobs connector needed `created_at` for incremental sync
- Commit message: *"These timestamps are critical for incremental sync in LazyJobs connector"*
- Without these fields, LazyJobs couldn't determine when jobs were added

**The bug introduced:**
- Added fields WITHOUT `omitempty` tag
- Connectors don't set these fields → Go zero time (`0001-01-01`)
- Zero time marshaled to JSON → `"created_at": "0001-01-01T00:00:00Z"`
- Sent to Supabase → Overrides `DEFAULT NOW()`

### ⚠️ **Nov 26, 01:15 (Commit 5122ebf): is_active Issue**

A separate but related issue:
- Added `is_active` column to database
- Go bool defaults to `false`
- Jobs inserted with `is_active = false`
- Jobs immediately hidden from API

### 🔧 **Nov 27, 23:58 (Commit 8e4b28e): is_active Fixed**

Fixed the `is_active` issue:
```go
job := models.JobPost{
    // ...
    IsActive: true,  // ✅ Explicitly set
}
```

**But `CreatedAt`/`UpdatedAt` still broken!**

### ❌ **Nov 28, 04:00: Zero Timestamps**

First sync after redeployment:
- 12 jobs inserted with `created_at = 0001-01-01`
- Jobs invisible to date queries
- LazyJobs connector couldn't find them

### ✅ **Nov 28, 12:49: Final Fix**

Added `omitempty` to fix the root cause:
```go
CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
```

---

## Root Cause Analysis

### Why It Worked Before Nov 26

**The struct didn't have CreatedAt/UpdatedAt fields:**
```go
// BEFORE (Nov 13-25)
type JobPost struct {
    ID          string
    Title       string
    // ...
    Source      string
    Fields      map[string]interface{}
    // NO CreatedAt
    // NO UpdatedAt
}
```

**What happened during insert:**
1. Connector creates JobPost (no timestamp fields)
2. Marshal to JSON → No `created_at`/`updated_at` keys
3. POST to Supabase → `{"id": "...", "title": "..."}`
4. Database sees missing fields → Uses `DEFAULT NOW()`
5. ✅ Correct timestamps!

### Why It Broke After Nov 26

**The struct gained CreatedAt/UpdatedAt WITHOUT omitempty:**
```go
// AFTER (Nov 26+)
type JobPost struct {
    ID          string
    Title       string
    // ...
    CreatedAt   time.Time `json:"created_at" db:"created_at"`  // ⚠️ NO omitempty
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`  // ⚠️ NO omitempty
}
```

**What happened during insert:**
1. Connector creates JobPost (timestamp fields exist but not set)
2. Fields default to Go zero time: `time.Time{}` = `0001-01-01`
3. Marshal to JSON → `{"created_at": "0001-01-01T00:00:00Z"}`
4. POST to Supabase with explicit zero time
5. Database receives explicit value → Ignores `DEFAULT NOW()`
6. ❌ Zero timestamps stored!

---

## The Go Zero Value Problem

### Go's Zero Values
```go
var t time.Time
fmt.Println(t.IsZero())  // true
fmt.Println(t)           // 0001-01-01 00:00:00 +0000 UTC
```

### JSON Marshaling Without omitempty
```go
type Job struct {
    Title     string    `json:"title"`
    CreatedAt time.Time `json:"created_at"`  // NO omitempty
}

job := Job{Title: "Developer"}
// CreatedAt is zero value

json, _ := json.Marshal(job)
fmt.Println(string(json))
// Output: {"title":"Developer","created_at":"0001-01-01T00:00:00Z"}
//                                           ^^^^^^^^^^^^^^^^^^^^^^^^
//                                           Zero time is included!
```

### JSON Marshaling WITH omitempty
```go
type Job struct {
    Title     string    `json:"title"`
    CreatedAt time.Time `json:"created_at,omitempty"`  // WITH omitempty
}

job := Job{Title: "Developer"}
// CreatedAt is zero value

json, _ := json.Marshal(job)
fmt.Println(string(json))
// Output: {"title":"Developer"}
//         No created_at field!
```

---

## Why It Only Affected Recent Jobs

### Database State

**Jobs before Nov 26:**
```sql
SELECT id, created_at FROM job_posts 
WHERE created_at < '2025-11-26'
LIMIT 3;

-- Results:
-- remotive-2074545 | 2025-11-13 06:00:01.098417+00  ✅ Correct
-- remotive-2074152 | 2025-11-13 06:00:01.12498+00   ✅ Correct
-- remotive-2074151 | 2025-11-13 06:00:01.14716+00   ✅ Correct
```

**Jobs after Nov 26 (before fix):**
```sql
SELECT id, created_at FROM job_posts 
WHERE id IN ('af-30315109', 'af-30314831', 'af-30314820');

-- Results:
-- af-30315109 | 0001-01-01 00:00:00+00  ❌ Zero time
-- af-30314831 | 0001-01-01 00:00:00+00  ❌ Zero time
-- af-30314820 | 0001-01-01 00:00:00+00  ❌ Zero time
```

### Why Only 20 Jobs Affected

The bug existed from Nov 26, but:
1. **Nov 26-27:** Plugins might not have been redeployed immediately
2. **Nov 27 23:58:** Redeployed with `IsActive: true` fix
3. **Nov 28 04:00:** First sync with new code → 12 jobs with zero timestamps
4. **Nov 28 12:49:** Bug discovered and fixed

**Total affected:** 20 jobs (from various syncs between Nov 26-28)

---

## Lessons Learned

### 1. Always Use omitempty for Optional Fields
```go
// GOOD ✅
CreatedAt time.Time `json:"created_at,omitempty"`

// BAD ❌
CreatedAt time.Time `json:"created_at"`
```

### 2. Test Database Defaults
When adding fields that rely on database defaults:
```go
// Test that zero values don't override defaults
job := models.JobPost{Title: "Test"}
// Don't set CreatedAt
store.CreateJob(&job)
// Verify CreatedAt is set by database
```

### 3. Incremental Changes Are Risky
The change was well-intentioned:
- ✅ Goal: Enable LazyJobs incremental sync
- ❌ Side effect: Broke timestamp handling

**Better approach:**
1. Add fields WITH `omitempty` from the start
2. Test in staging environment
3. Monitor first production sync

### 4. Go Zero Values Are Dangerous
Go's zero values can override database defaults:
- `bool` → `false` (overrides `DEFAULT true`)
- `time.Time` → `0001-01-01` (overrides `DEFAULT NOW()`)
- `int` → `0` (overrides `DEFAULT 1`)

**Solution:** Use `omitempty` or pointers for optional fields

---

## Summary

**Why it worked before:**
- Fields didn't exist → Couldn't be set → Database defaults worked

**Why it broke:**
- Fields added without `omitempty` → Zero values sent → Database defaults overridden

**Why it only affected recent jobs:**
- Bug introduced Nov 26 → Plugins redeployed Nov 27 → First broken sync Nov 28

**How we fixed it:**
- Added `omitempty` → Zero values omitted → Database defaults work again

**Affected period:**
- Nov 26 00:36 - Nov 28 12:49 (approximately 60 hours)
- Only 20 jobs affected (out of 813 total)
- All fixed with migration script

---

## Prevention

### Code Review Checklist
When adding fields to database models:
- [ ] Does field rely on database default?
- [ ] Is field optional?
- [ ] Does field have `omitempty` tag?
- [ ] Are zero values tested?
- [ ] Is there a migration plan?

### Testing Checklist
- [ ] Test with zero values
- [ ] Test with explicit values
- [ ] Test with missing fields
- [ ] Verify database defaults work
- [ ] Check JSON marshaling output

### Deployment Checklist
- [ ] Test in staging first
- [ ] Monitor first production sync
- [ ] Have rollback plan ready
- [ ] Document breaking changes
- [ ] Alert team of changes
