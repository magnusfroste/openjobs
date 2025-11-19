-- Migration 003: Row Level Security Policies
-- Secures job_posts table so only OpenJobs API can write

-- Enable RLS on job_posts
ALTER TABLE job_posts ENABLE ROW LEVEL SECURITY;

-- Policy 1: Everyone can READ active jobs (public API)
CREATE POLICY "Enable read access for all users" ON job_posts
FOR SELECT
USING (is_active = true);

-- Policy 2: Only service_role can INSERT (OpenJobs API uses service_role key)
CREATE POLICY "Enable insert for service_role only" ON job_posts
FOR INSERT
TO service_role
WITH CHECK (true);

-- Policy 3: Only service_role can UPDATE
CREATE POLICY "Enable update for service_role only" ON job_posts
FOR UPDATE
TO service_role
USING (true);

-- Policy 4: Only service_role can DELETE
CREATE POLICY "Enable delete for service_role only" ON job_posts
FOR DELETE
TO service_role
USING (true);

-- Note: This means OpenJobs API must use SERVICE_ROLE_KEY (not ANON_KEY) for write operations
-- The API validates X-API-Key before writing, then uses service_role to actually insert

-- Enable RLS on companies table
ALTER TABLE companies ENABLE ROW LEVEL SECURITY;

-- Policy: Allow anon to read for API key validation (OpenJobs API needs this)
CREATE POLICY "Enable read for API key validation" ON companies
FOR SELECT
TO anon
USING (true);

-- Policy: Only authenticated users can insert their own company (registration)
CREATE POLICY "Enable insert for authenticated users" ON companies
FOR INSERT
TO authenticated
WITH CHECK (auth.uid() = user_id);

-- Policy: Users can read their own company
CREATE POLICY "Enable read for own company" ON companies
FOR SELECT
TO authenticated
USING (auth.uid() = user_id);

-- Policy: Users can update their own company
CREATE POLICY "Enable update for own company" ON companies
FOR UPDATE
TO authenticated
USING (auth.uid() = user_id);
