-- Fix jobs with zero timestamps (0001-01-01)
-- These jobs were inserted before the pointer fix was applied
-- Set their created_at and updated_at to their posted_date

UPDATE job_posts 
SET 
  created_at = posted_date,
  updated_at = posted_date
WHERE created_at = '0001-01-01 00:00:00+00';

-- Verify the fix
SELECT 
    COUNT(*) as total_jobs,
    COUNT(CASE WHEN created_at = '0001-01-01 00:00:00+00' THEN 1 END) as zero_timestamp_jobs,
    COUNT(CASE WHEN created_at >= NOW() - INTERVAL '1 day' THEN 1 END) as jobs_last_24h
FROM job_posts;
