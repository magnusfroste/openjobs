# Testing Database Connection

After migrating to the new cloud Supabase, use these tests to verify OpenJobs can read and write to the database.

## Prerequisites

1. **Update .env file** with new Supabase credentials:
```bash
SUPABASE_URL=https://cmpnqpdxhmecptcbffmw.supabase.co
SUPABASE_ANON_KEY=your-anon-key-here
```

2. **Run database migrations** in Supabase SQL Editor:
```sql
-- Run these in order:
migrations/001_create_job_posts.sql
migrations/002_enhanced_schema.sql
```

3. **Create companies table** (for API key validation):
```sql
-- See OpenJobs_Web README for full SQL
CREATE TABLE IF NOT EXISTS public.companies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  api_key VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Test Methods

### Option 1: Shell Script (Quick Test)

Tests basic HTTP connectivity to Supabase REST API.

```bash
# Make executable
chmod +x test_supabase_connection.sh

# Run tests
./test_supabase_connection.sh
```

**What it tests:**
- ✅ Supabase health check
- ✅ Read from job_posts table
- ✅ Write to job_posts table
- ✅ Update job_posts table
- ✅ Delete from job_posts table
- ✅ Read from companies table

### Option 2: Go Test Program (Full Test)

Tests using actual OpenJobs code (JobStore).

```bash
# Run from OpenJobs root directory
go run cmd/test-db/main.go
```

**What it tests:**
- ✅ Read jobs using GetAllJobs()
- ✅ Get total job count
- ✅ Create job using CreateJob()
- ✅ Read job back using GetJob()
- ✅ Update job using UpdateJob()
- ✅ Delete job using DeleteJob()
- ✅ Get remote job count

## Expected Output

### Successful Test Output:
```
🧪 Testing OpenJobs → Supabase Connection
==========================================

📍 Supabase URL: https://cmpnqpdxhmecptcbffmw.supabase.co

Test 1: Read Jobs from Database
--------------------------------
✅ Successfully read 5 jobs
  1. Senior Developer - Tech AB (ID: af-12345)
  2. React Engineer - Startup Inc (ID: ro-67890)
  ...

Test 2: Get Total Job Count
---------------------------
✅ Total active jobs in database: 333

Test 3: Create Test Job
-----------------------
✅ Successfully created test job: test-1732012345

Test 4: Read Test Job Back
--------------------------
✅ Successfully read test job back
  Title: Test Job - Connection Check
  Company: OpenJobs Test
  Active: false

Test 5: Update Test Job
-----------------------
✅ Successfully updated test job

Test 6: Cleanup (Delete Test Job)
---------------------------------
✅ Successfully deleted test job
🧹 Test job cleaned up

Test 7: Get Remote Job Count
----------------------------
✅ Total remote jobs: 168

==========================================
📊 Test Summary
==========================================
✅ All database operations working!
```

### Common Errors

**Error: Missing SUPABASE_URL**
```
❌ Missing SUPABASE_URL or SUPABASE_ANON_KEY
```
**Fix:** Update your `.env` file with the new cloud Supabase credentials.

**Error: Failed to read jobs (HTTP 404)**
```
❌ Failed to read jobs: supabase error 404
```
**Fix:** Run the database migrations. The `job_posts` table doesn't exist yet.

**Error: Failed to write (HTTP 401)**
```
❌ Failed to create test job: supabase error 401
```
**Fix:** Check that your `SUPABASE_ANON_KEY` is correct and has the right permissions.

**Error: Companies table check failed**
```
⚠️  Companies table check (HTTP 404)
   Note: This is expected if table doesn't exist yet
```
**Fix:** This is expected if you haven't created the companies table yet. Run the SQL from the prerequisites section.

## Next Steps After Successful Tests

1. **Start OpenJobs API:**
```bash
go run cmd/openjobs/main.go
```

2. **Test API health:**
```bash
curl http://localhost:8080/health
```

3. **Test jobs endpoint:**
```bash
curl http://localhost:8080/jobs?limit=5
```

4. **Trigger manual sync:**
```bash
curl -X POST http://localhost:8080/sync/manual
```

5. **Deploy to production:**
```bash
# Rebuild Docker image
docker build -t openjobs:latest .

# Or push to Easypanel
git push
```

## Troubleshooting

### Connection Timeout
- Check that Supabase URL is correct
- Verify your internet connection
- Check Supabase project is not paused

### Permission Errors
- Verify SUPABASE_ANON_KEY is correct
- Check RLS policies in Supabase
- Ensure anon role has proper permissions

### Data Not Appearing
- Check `is_active=true` filter
- Verify jobs have `posted_date` set
- Look at Supabase logs for errors

## Clean Up Test Data

If tests fail and leave test jobs in database:

```sql
-- Delete all test jobs
DELETE FROM job_posts WHERE id LIKE 'test-%';
```

Or via API:
```bash
curl -X DELETE \
  -H "apikey: $SUPABASE_ANON_KEY" \
  -H "Authorization: Bearer $SUPABASE_ANON_KEY" \
  "${SUPABASE_URL}/rest/v1/job_posts?id=like.test-*"
```
