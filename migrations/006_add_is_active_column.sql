-- Add is_active column to job_posts table
-- This column is used to filter active/expired jobs
-- Default to true for all existing jobs

ALTER TABLE job_posts 
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true NOT NULL;

-- Create index for efficient filtering
CREATE INDEX IF NOT EXISTS idx_job_posts_is_active 
ON job_posts (is_active) 
WHERE is_active = true;

-- Update existing jobs to be active
UPDATE job_posts 
SET is_active = true 
WHERE is_active IS NULL;

-- Add comment explaining the field
COMMENT ON COLUMN job_posts.is_active IS 'Whether the job posting is currently active (not expired or closed)';

-- Verify the changes
SELECT 
    COUNT(*) as total_jobs,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_jobs,
    COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_jobs
FROM job_posts;
