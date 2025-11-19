-- Migration 005: Update job_analytics materialized view for new schema
-- This replaces the old view that used plugin_source, location_country, etc.

-- Drop old materialized view if exists
DROP MATERIALIZED VIEW IF EXISTS job_analytics CASCADE;

-- Create new materialized view with current schema
CREATE MATERIALIZED VIEW job_analytics AS
SELECT
    source,  -- New column name (was plugin_source)
    
    -- Job counts
    COUNT(*) as total_jobs,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_jobs,
    COUNT(CASE WHEN is_remote = true THEN 1 END) as remote_jobs,
    
    -- Salary analytics
    ROUND(AVG(salary_min)) as avg_min_salary,
    ROUND(AVG(salary_max)) as avg_max_salary,
    COUNT(CASE WHEN salary_min IS NOT NULL OR salary_max IS NOT NULL THEN 1 END) as jobs_with_salary,
    
    -- Geographic spread (from fields JSONB)
    COUNT(DISTINCT fields->>'country') FILTER (WHERE fields->>'country' IS NOT NULL) as countries_covered,
    
    -- Employment type distribution
    COUNT(CASE WHEN employment_type = 'full-time' THEN 1 END) as fulltime_jobs,
    COUNT(CASE WHEN employment_type = 'part-time' THEN 1 END) as parttime_jobs,
    COUNT(CASE WHEN employment_type = 'contract' THEN 1 END) as contract_jobs,
    
    -- Freshness and sync metrics
    MAX(created_at) as latest_sync,
    MIN(created_at) as first_sync,
    MAX(posted_date) as latest_job_posted,
    
    -- Data quality metrics
    COUNT(CASE WHEN url IS NOT NULL THEN 1 END) as jobs_with_url,
    COUNT(CASE WHEN requirements IS NOT NULL AND array_length(requirements, 1) > 0 THEN 1 END) as jobs_with_requirements,
    COUNT(CASE WHEN benefits IS NOT NULL AND array_length(benefits, 1) > 0 THEN 1 END) as jobs_with_benefits

FROM job_posts
GROUP BY source;

-- Create index for faster queries
CREATE INDEX idx_job_analytics_source ON job_analytics(source);

-- Grant permissions
GRANT SELECT ON job_analytics TO authenticated;
GRANT SELECT ON job_analytics TO anon;

-- SECURITY NOTE: Materialized views do NOT support Row-Level Security (RLS)
-- This means ALL data in job_analytics is accessible to anyone with anon key
-- This is acceptable because:
-- 1. Analytics data is meant to be public (job counts, averages, etc.)
-- 2. No sensitive user data is included
-- 3. Access is through /analytics endpoint which is intentionally public
-- 
-- If you need to restrict analytics access in the future:
-- - Remove GRANT to anon
-- - Use service_role key in GetAnalyticsBySource()
-- - Add authentication to /analytics endpoint

-- Function to refresh materialized view
CREATE OR REPLACE FUNCTION refresh_job_analytics()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW job_analytics;
END;
$$ LANGUAGE plpgsql;

-- Function to get analytics summary (all sources combined)
CREATE OR REPLACE FUNCTION get_job_analytics_summary()
RETURNS TABLE (
    total_jobs bigint,
    active_jobs bigint,
    sources_count bigint,
    remote_percentage decimal,
    avg_salary_range int
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        SUM(ja.total_jobs) as total_jobs,
        SUM(ja.active_jobs) as active_jobs,
        COUNT(*) as sources_count,
        ROUND((SUM(ja.remote_jobs)::decimal / NULLIF(SUM(ja.total_jobs), 0)) * 100, 2) as remote_percentage,
        ROUND(AVG(ja.avg_max_salary - ja.avg_min_salary)) as avg_salary_range
    FROM job_analytics ja;
END;
$$ LANGUAGE plpgsql;

-- Function to get analytics by source
CREATE OR REPLACE FUNCTION get_job_analytics_by_source()
RETURNS TABLE (
    source text,
    total_jobs bigint,
    active_jobs bigint,
    remote_jobs bigint,
    countries_covered bigint,
    avg_min_salary numeric,
    avg_max_salary numeric,
    latest_sync timestamp with time zone
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ja.source::text,
        ja.total_jobs,
        ja.active_jobs,
        ja.remote_jobs,
        ja.countries_covered,
        ja.avg_min_salary,
        ja.avg_max_salary,
        ja.latest_sync
    FROM job_analytics ja
    ORDER BY ja.total_jobs DESC;
END;
$$ LANGUAGE plpgsql;

-- Comments for documentation
COMMENT ON MATERIALIZED VIEW job_analytics IS 'Pre-computed analytics for job platform dashboard (refresh after sync)';
COMMENT ON FUNCTION refresh_job_analytics() IS 'Refreshes the job_analytics materialized view with latest data';
COMMENT ON FUNCTION get_job_analytics_summary() IS 'Returns high-level analytics summary across all sources';
COMMENT ON FUNCTION get_job_analytics_by_source() IS 'Returns detailed analytics broken down by source';

-- Initial refresh
REFRESH MATERIALIZED VIEW job_analytics;
