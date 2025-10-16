-- Add fields for better job matching
-- These fields enable salary filtering, remote work filtering, and direct application links

-- Add new columns
ALTER TABLE job_posts 
ADD COLUMN IF NOT EXISTS salary_min INTEGER,
ADD COLUMN IF NOT EXISTS salary_max INTEGER,
ADD COLUMN IF NOT EXISTS salary_currency VARCHAR(10) DEFAULT 'USD',
ADD COLUMN IF NOT EXISTS is_remote BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS url TEXT;

-- Create indexes for efficient filtering
CREATE INDEX IF NOT EXISTS idx_job_posts_salary_min 
ON job_posts (salary_min) 
WHERE salary_min IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_job_posts_is_remote 
ON job_posts (is_remote) 
WHERE is_remote = true;

CREATE INDEX IF NOT EXISTS idx_job_posts_url 
ON job_posts (url) 
WHERE url IS NOT NULL;

-- Backfill is_remote from existing data
-- Check location and description for remote keywords
UPDATE job_posts 
SET is_remote = true 
WHERE 
    is_remote = false
    AND (
        LOWER(location) LIKE '%remote%' 
        OR LOWER(location) LIKE '%distans%'
        OR LOWER(location) LIKE '%hemarbete%'
        OR LOWER(location) LIKE '%fjärr%'
        OR LOWER(description) LIKE '%remote%'
        OR LOWER(description) LIKE '%distans%'
        OR fields->>'source' IN ('remoteok', 'remotive')
        OR fields->>'connector' IN ('remoteok', 'remotive')
    );

-- Backfill URL from fields.source_url
UPDATE job_posts 
SET url = fields->>'source_url'
WHERE 
    url IS NULL 
    AND fields->>'source_url' IS NOT NULL 
    AND fields->>'source_url' != '';

-- Add comment explaining the new fields
COMMENT ON COLUMN job_posts.salary_min IS 'Minimum salary (parsed from salary string)';
COMMENT ON COLUMN job_posts.salary_max IS 'Maximum salary (parsed from salary string)';
COMMENT ON COLUMN job_posts.salary_currency IS 'Salary currency code (e.g., USD, SEK, EUR)';
COMMENT ON COLUMN job_posts.is_remote IS 'Whether the job allows remote work';
COMMENT ON COLUMN job_posts.url IS 'Direct URL to job application page';

-- Verify the changes
SELECT 
    COUNT(*) as total_jobs,
    COUNT(salary_min) as jobs_with_salary_min,
    COUNT(salary_max) as jobs_with_salary_max,
    COUNT(CASE WHEN is_remote = true THEN 1 END) as remote_jobs,
    COUNT(url) as jobs_with_url
FROM job_posts;
