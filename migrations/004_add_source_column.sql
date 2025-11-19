-- Migration 004: Add source column for better analytics
-- This separates ingestion source from metadata

-- Add source column
ALTER TABLE job_posts 
ADD COLUMN source VARCHAR(50);

-- Create index for fast queries
CREATE INDEX idx_job_posts_source ON job_posts(source);

-- Migrate existing data from fields->>'source' to source column
UPDATE job_posts 
SET source = fields->>'source'
WHERE fields->>'source' IS NOT NULL;

-- Set default for rows without source
UPDATE job_posts 
SET source = 'unknown'
WHERE source IS NULL;

-- Make it NOT NULL with default
ALTER TABLE job_posts 
ALTER COLUMN source SET NOT NULL,
ALTER COLUMN source SET DEFAULT 'openjobs-api';

-- No CHECK constraint - allows flexibility for new sources without migrations
-- Instead, we rely on:
-- 1. Application-level validation in API
-- 2. Documentation in SOURCE_GUIDE.md
-- 3. Analytics queries to detect unusual sources

-- Add comment
COMMENT ON COLUMN job_posts.source IS 'Primary source of the job posting (ingestion channel)';

-- Optional: Migrate legacy 'web-form' to 'openjobs-web'
UPDATE job_posts 
SET source = 'openjobs-web'
WHERE source = 'web-form';
