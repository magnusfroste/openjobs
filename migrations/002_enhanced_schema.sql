-- Enhanced schema migration for analytics and plugin diversity support
-- Run this after the base migration (001_create_job_posts.sql)

-- Add analytics columns to core job_posts table
ALTER TABLE job_posts
ADD COLUMN IF NOT EXISTS plugin_source VARCHAR(100) DEFAULT 'unknown',
ADD COLUMN IF NOT EXISTS location_country VARCHAR(100),
ADD COLUMN IF NOT EXISTS job_category VARCHAR(100),
ADD COLUMN IF NOT EXISTS salary_numeric INT[],
ADD COLUMN IF NOT EXISTS tags TEXT[],
ADD COLUMN IF NOT EXISTS remote_work BOOLEAN DEFAULT FALSE;

-- Create plugin-specific data table for rich metadata
CREATE TABLE IF NOT EXISTS job_posts_plugin_data (
    job_id VARCHAR(255) NOT NULL,
    plugin_source VARCHAR(100) NOT NULL,

    -- Plugin-specific rich metadata (indexed for analytics)
    structured_data JSONB,

    -- Raw API response backup
    raw_data JSONB,

    -- Metadata tracking
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (job_id, plugin_source),
    FOREIGN KEY (job_id) REFERENCES job_posts(id) ON DELETE CASCADE
);

-- Performance indexes for analytics
CREATE INDEX IF NOT EXISTS idx_plugin_data_source ON job_posts_plugin_data (plugin_source);
CREATE INDEX IF NOT EXISTS idx_structured_data ON job_posts_plugin_data USING GIN (structured_data);
CREATE INDEX IF NOT EXISTS idx_location_country ON job_posts (location_country);
CREATE INDEX IF NOT EXISTS idx_job_category ON job_posts (job_category);
CREATE INDEX IF NOT EXISTS idx_plugin_source ON job_posts (plugin_source);
CREATE INDEX IF NOT EXISTS idx_created_at ON job_posts (created_at);
CREATE INDEX IF NOT EXISTS idx_remote_work ON job_posts (remote_work);

-- Materialized view for high-performance analytics
CREATE MATERIALIZED VIEW IF NOT EXISTS job_analytics AS
SELECT
    jp.plugin_source,

    -- Job counts and geographic spread
    COUNT(*) as total_jobs,
    COUNT(DISTINCT jp.location_country) as countries_covered,

    -- Employment type distribution
    COUNT(CASE WHEN jp.employment_type = 'Full-time' THEN 1 END) as fulltime_jobs,
    COUNT(CASE WHEN jp.employment_type = 'Part-time' THEN 1 END) as parttime_jobs,
    COUNT(CASE WHEN jp.remote_work = TRUE THEN 1 END) as remote_jobs,

    -- Salary analytics (from structured data where available)
    ROUND(AVG((jpd.structured_data->>'salary_min')::float)) as avg_min_salary,
    ROUND(AVG((jpd.structured_data->>'salary_max')::float)) as avg_max_salary,

    -- Category diversity (plugin-specific)
    COUNT(DISTINCT CASE
        WHEN jp.plugin_source = 'arbetsformedlingen' THEN jpd.structured_data->>'occupation_group'
        WHEN jp.plugin_source = 'eures' THEN jpd.structured_data->>'job_category'
        WHEN jp.plugin_source = 'remotive' THEN jpd.structured_data->>'category'
        ELSE jp.job_category
    END) as categories_count,

    -- Freshness and sync metrics
    MAX(jp.created_at) as latest_job,
    MIN(jp.created_at) as first_job,
    ROUND(EXTRACT(EPOCH FROM (MAX(jp.created_at) - MIN(jp.created_at))) / 3600, 2) as hours_active,

    -- Metadata richness tracking
    COUNT(*) FILTER (WHERE jpd.structured_data IS NOT NULL) as jobs_with_metadata,
    AVG(array_length(jp.tags, 1)) FILTER (WHERE array_length(jp.tags, 1) IS NOT NULL) as avg_tags_per_job

FROM job_posts jp
LEFT JOIN job_posts_plugin_data jpd ON jp.id = jpd.job_id AND jp.plugin_source = jpd.plugin_source
WHERE jp.plugin_source IN ('arbetsformedlingen', 'eures', 'remotive')
GROUP BY jp.plugin_source;

-- Grant permissions for analytics queries
GRANT SELECT ON job_analytics TO authenticated;
GRANT SELECT ON job_posts_plugin_data TO authenticated;

-- Function to refresh materialized view (run periodically or after bulk imports)
CREATE OR REPLACE FUNCTION refresh_job_analytics()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW job_analytics;
END;
$$ LANGUAGE plpgsql;

-- Function to get analytics summary
CREATE OR REPLACE FUNCTION get_job_analytics_summary()
RETURNS TABLE (
    total_jobs bigint,
    sources_count bigint,
    countries_covered bigint,
    avg_salary_range int,
    remote_percentage decimal
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        SUM(ja.total_jobs) as total_jobs,
        COUNT(*) as sources_count,
        SUM(ja.countries_covered) as countries_covered,
        ROUND(AVG(ja.avg_max_salary - ja.avg_min_salary)) as avg_salary_range,
        ROUND((SUM(ja.remote_jobs)::decimal / SUM(ja.total_jobs)) * 100, 2) as remote_percentage
    FROM job_analytics ja;
END;
$$ LANGUAGE plpgsql;

-- Comments for documentation
COMMENT ON TABLE job_posts_plugin_data IS 'Stores rich plugin-specific metadata and raw API responses for analytics';
COMMENT ON MATERIALIZED VIEW job_analytics IS 'Pre-computed analytics for job platform dashboard (refresh periodically)';
COMMENT ON FUNCTION refresh_job_analytics() IS 'Refreshes the job_analytics materialized view with latest data';
COMMENT ON FUNCTION get_job_analytics_summary() IS 'Returns high-level analytics summary across all sources';
