CREATE INDEX IF NOT EXISTS idx_events_site_timestamp 
ON events (site_id, timestamp);

CREATE INDEX IF NOT EXISTS idx_events_path 
ON events (path);
