-- Fix jobs with zero timestamps (0001-01-01) from connector bug
-- These jobs were inserted but with Go's zero time value instead of NOW()

-- Update jobs with zero created_at to use posted_date as fallback
UPDATE job_posts 
SET 
    created_at = COALESCE(
        CASE 
            WHEN created_at < '2000-01-01' THEN posted_date 
            ELSE created_at 
        END,
        NOW()
    ),
    updated_at = COALESCE(
        CASE 
            WHEN updated_at < '2000-01-01' THEN posted_date 
            ELSE updated_at 
        END,
        NOW()
    )
WHERE created_at < '2000-01-01' OR updated_at < '2000-01-01';

-- Verify the fix
SELECT 
    COUNT(*) as total_jobs,
    COUNT(CASE WHEN created_at < '2000-01-01' THEN 1 END) as jobs_with_zero_created_at,
    COUNT(CASE WHEN updated_at < '2000-01-01' THEN 1 END) as jobs_with_zero_updated_at,
    MIN(created_at) as earliest_created_at,
    MAX(created_at) as latest_created_at
FROM job_posts;
