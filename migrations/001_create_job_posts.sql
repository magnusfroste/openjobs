-- Create job_posts table
CREATE TABLE IF NOT EXISTS job_posts (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    company VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255),
    salary VARCHAR(255),
    employment_type VARCHAR(100),
    experience_level VARCHAR(100),
    posted_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_date TIMESTAMP WITH TIME ZONE,
    requirements TEXT[],
    benefits TEXT[],
    fields JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on posted_date for efficient ordering
CREATE INDEX IF NOT EXISTS idx_job_posts_posted_date ON job_posts (posted_date DESC);

-- Create index on location for filtering
CREATE INDEX IF NOT EXISTS idx_job_posts_location ON job_posts (location);

-- Create index on employment_type for filtering
CREATE INDEX IF NOT EXISTS idx_job_posts_employment_type ON job_posts (employment_type);

-- Create GIN index on fields for JSONB queries
CREATE INDEX IF NOT EXISTS idx_job_posts_fields ON job_posts USING GIN (fields);

-- Create plugins table
CREATE TABLE IF NOT EXISTS plugins (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    source VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'inactive',
    last_run TIMESTAMP WITH TIME ZONE,
    next_run TIMESTAMP WITH TIME ZONE,
    description TEXT,
    config JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);