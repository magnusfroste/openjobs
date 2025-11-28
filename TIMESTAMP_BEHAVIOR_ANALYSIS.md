# Timestamp Behavior Analysis: CreatedAt/UpdatedAt with omitempty

## Question
How does adding `omitempty` to `CreatedAt`/`UpdatedAt` affect different insertion methods?

## TL;DR
✅ **The fix is safe for all insertion methods:**
- **Connectors:** ✅ Fixed - now use database defaults
- **API (companies):** ✅ Still works - can set timestamps explicitly
- **Direct Supabase:** ✅ Unaffected - database defaults work

## Detailed Analysis

### The Fix
```go
// BEFORE
CreatedAt time.Time `json:"created_at" db:"created_at"`
UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

// AFTER
CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
```

### How `omitempty` Works

When marshaling a struct to JSON:
- **Without `omitempty`:** Zero values are included → `{"created_at": "0001-01-01T00:00:00Z"}`
- **With `omitempty`:** Zero values are excluded → `{}` (field not present)

When a field is **not present** in JSON, Supabase uses the database default (`DEFAULT NOW()`).

---

## Insertion Method #1: Plugin Connectors (FIXED ✅)

### Code Example
```go
// connectors/arbetsformedlingen/connector.go
job := models.JobPost{
    ID:          "af-12345",
    Title:       "Developer",
    Company:     "ACME",
    PostedDate:  time.Now(),
    // CreatedAt NOT SET (zero value)
    // UpdatedAt NOT SET (zero value)
}
store.CreateJob(&job)
```

### Behavior

**BEFORE Fix (BROKEN ❌):**
```json
{
  "id": "af-12345",
  "title": "Developer",
  "created_at": "0001-01-01T00:00:00Z",  // ❌ Zero time sent
  "updated_at": "0001-01-01T00:00:00Z"   // ❌ Zero time sent
}
```
Result: Database receives zero time, overrides `DEFAULT NOW()`

**AFTER Fix (WORKING ✅):**
```json
{
  "id": "af-12345",
  "title": "Developer"
  // created_at NOT PRESENT (omitted)
  // updated_at NOT PRESENT (omitted)
}
```
Result: Database uses `DEFAULT NOW()` → correct timestamps

### Impact
✅ **All connectors now work correctly:**
- arbetsformedlingen
- eures
- remotive
- remoteok
- indeed
- indeed-scraper
- indeed-chrome
- jooble
- offentligajobb

---

## Insertion Method #2: OpenJobs API (POST /jobs) - STILL WORKS ✅

### Code Example
```go
// internal/api/handlers.go - CreateJob()
func (s *Server) CreateJob(w http.ResponseWriter, r *http.Request) {
    var job models.JobPost
    json.NewDecoder(r.Body).Decode(&job)
    
    // Set timestamps EXPLICITLY
    now := time.Now()
    if job.PostedDate.IsZero() {
        job.PostedDate = now  // ✅ Explicitly set
    }
    
    // CreatedAt and UpdatedAt NOT set (zero values)
    // This is INTENTIONAL - let database handle it
    
    s.jobStore.CreateJob(&job)
}
```

### Behavior

**Scenario A: Company doesn't send timestamps (MOST COMMON)**
```bash
curl -X POST https://api.openjobs.ink/jobs \
  -H "X-API-Key: company-key" \
  -d '{
    "title": "Senior Developer",
    "company": "TechCorp",
    "description": "..."
  }'
```

JSON sent to Supabase:
```json
{
  "id": "web-uuid",
  "title": "Senior Developer",
  "company": "TechCorp",
  "posted_date": "2025-11-28T12:00:00Z"
  // created_at NOT PRESENT → Database uses NOW()
  // updated_at NOT PRESENT → Database uses NOW()
}
```
✅ **Result:** Correct timestamps from database

**Scenario B: Company explicitly sends timestamps (ADVANCED)**
```bash
curl -X POST https://api.openjobs.ink/jobs \
  -H "X-API-Key: company-key" \
  -d '{
    "title": "Senior Developer",
    "company": "TechCorp",
    "description": "...",
    "created_at": "2025-11-28T10:00:00Z",
    "updated_at": "2025-11-28T10:00:00Z"
  }'
```

JSON sent to Supabase:
```json
{
  "id": "web-uuid",
  "title": "Senior Developer",
  "created_at": "2025-11-28T10:00:00Z",  // ✅ Explicit value honored
  "updated_at": "2025-11-28T10:00:00Z"   // ✅ Explicit value honored
}
```
✅ **Result:** Company's timestamps are used

### Impact
✅ **API still works for all scenarios:**
- Default case: Database handles timestamps
- Advanced case: Company can override timestamps
- No breaking changes

---

## Insertion Method #3: Direct Supabase Insert - UNAFFECTED ✅

### Example
```bash
curl -X POST "https://cmpnqpdxhmecptcbffmw.supabase.co/rest/v1/job_posts" \
  -H "apikey: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "custom-123",
    "title": "Job Title",
    "company": "Company"
  }'
```

### Behavior
✅ **Unaffected** - This bypasses Go entirely, goes straight to Supabase
- Database `DEFAULT NOW()` constraint applies
- Works exactly as before

---

## Edge Cases

### Edge Case 1: Updating Jobs (PATCH)
```go
// pkg/storage/job.go - UpdateJob()
func (js *JobStore) UpdateJob(job *models.JobPost) error {
    // When updating, we might want to preserve created_at
    // but update updated_at
}
```

**Current behavior:**
- If `UpdatedAt` is zero → omitted → database keeps old value
- If `UpdatedAt` is set → included → database uses new value

**Recommendation:** When updating jobs, explicitly set `UpdatedAt = time.Now()`

### Edge Case 2: Bulk Imports
If importing historical jobs with specific timestamps:
```go
job := models.JobPost{
    ID:        "historical-123",
    Title:     "Old Job",
    CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // ✅ Explicit
}
```
✅ **Works:** Non-zero timestamps are included in JSON

### Edge Case 3: Zero Time Detection
```go
if job.CreatedAt.IsZero() {
    // CreatedAt not set, will use database default
} else {
    // CreatedAt explicitly set, will be sent to database
}
```

---

## Database Schema (For Reference)

```sql
CREATE TABLE job_posts (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    company VARCHAR(255) NOT NULL,
    -- ...
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**Key points:**
- `NOT NULL` → Field is required
- `DEFAULT NOW()` → Used when field not provided
- If field IS provided → Default is ignored

---

## Summary Table

| Insertion Method | CreatedAt Behavior | Impact |
|-----------------|-------------------|---------|
| **Connectors** | Zero → Omitted → DB Default | ✅ FIXED |
| **API (no timestamp)** | Zero → Omitted → DB Default | ✅ WORKS |
| **API (with timestamp)** | Set → Included → Used | ✅ WORKS |
| **Direct Supabase** | N/A (bypasses Go) | ✅ UNAFFECTED |
| **Bulk Import (historical)** | Set → Included → Used | ✅ WORKS |

---

## Recommendations

### For Connector Developers
✅ **Do NOT set `CreatedAt`/`UpdatedAt`** - Let database handle it
```go
// GOOD ✅
job := models.JobPost{
    ID:    "af-123",
    Title: "Job",
    // CreatedAt not set
}

// BAD ❌ (but won't break anything)
job := models.JobPost{
    ID:        "af-123",
    Title:     "Job",
    CreatedAt: time.Now(), // Unnecessary
}
```

### For API Users (Companies)
✅ **Option 1 (Recommended):** Don't send timestamps
```json
{
  "title": "Job",
  "company": "ACME"
}
```

✅ **Option 2 (Advanced):** Send explicit timestamps
```json
{
  "title": "Job",
  "company": "ACME",
  "created_at": "2025-11-28T10:00:00Z"
}
```

### For Update Operations
⚠️ **Always set `UpdatedAt` explicitly when updating:**
```go
job.UpdatedAt = time.Now()
store.UpdateJob(&job)
```

---

## Testing

### Test Case 1: Connector Insert
```bash
# Run any connector
curl -X POST http://localhost:8081/sync

# Verify timestamps
curl "https://cmpnqpdxhmecptcbffmw.supabase.co/rest/v1/job_posts?select=id,created_at&order=created_at.desc&limit=1" \
  -H "apikey: YOUR_KEY"

# Expected: created_at is recent (not 0001-01-01)
```

### Test Case 2: API Insert (No Timestamp)
```bash
curl -X POST http://localhost:8080/jobs \
  -H "X-API-Key: test-key" \
  -d '{"title":"Test","company":"ACME","description":"Test"}'

# Verify timestamps are set correctly
```

### Test Case 3: API Insert (With Timestamp)
```bash
curl -X POST http://localhost:8080/jobs \
  -H "X-API-Key: test-key" \
  -d '{
    "title":"Test",
    "company":"ACME",
    "description":"Test",
    "created_at":"2025-01-01T00:00:00Z"
  }'

# Verify custom timestamp is honored
```

---

## Conclusion

✅ **The fix is safe and backward-compatible:**
- Connectors: Fixed the bug
- API: Still works for all use cases
- Direct Supabase: Unaffected
- No breaking changes

The `omitempty` tag is the correct solution because it:
1. Fixes the connector bug (zero times not sent)
2. Preserves API flexibility (explicit times still work)
3. Follows Go best practices
4. Aligns with database design (DEFAULT NOW())
