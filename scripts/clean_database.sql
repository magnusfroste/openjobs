-- Clean OpenJobs Database
-- Removes all test data for fresh start with new JobSearch API

-- Clear all job posts
DELETE FROM job_posts;

-- Clear all sync logs
DELETE FROM sync_logs;

-- Verify cleanup
SELECT 'job_posts' as table_name, COUNT(*) as remaining_rows FROM job_posts
UNION ALL
SELECT 'sync_logs' as table_name, COUNT(*) as remaining_rows FROM sync_logs;

-- Success message
SELECT 'OpenJobs database cleaned! Ready for fresh sync.' as status;
