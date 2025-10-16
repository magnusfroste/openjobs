-- Create sync logs table to track connector efficiency
-- This helps prevent over-polling of external APIs

CREATE TABLE IF NOT EXISTS sync_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_name VARCHAR(100) NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    jobs_fetched INTEGER NOT NULL DEFAULT 0,
    jobs_inserted INTEGER NOT NULL DEFAULT 0,
    jobs_duplicates INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'success', -- success, error, partial
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for fast queries by connector and time
CREATE INDEX IF NOT EXISTS idx_sync_logs_connector_time 
ON sync_logs (connector_name, started_at DESC);

-- Index for recent logs
CREATE INDEX IF NOT EXISTS idx_sync_logs_created_at 
ON sync_logs (created_at DESC);

-- Add comments
COMMENT ON TABLE sync_logs IS 'Tracks sync operations to monitor API efficiency and prevent over-polling';
COMMENT ON COLUMN sync_logs.jobs_fetched IS 'Total jobs fetched from external API';
COMMENT ON COLUMN sync_logs.jobs_inserted IS 'New jobs inserted (not duplicates)';
COMMENT ON COLUMN sync_logs.jobs_duplicates IS 'Duplicate jobs skipped';
COMMENT ON COLUMN sync_logs.status IS 'Sync status: success, error, or partial';

-- Verify the table
SELECT 
    'sync_logs table created' as message,
    COUNT(*) as initial_count
FROM sync_logs;
