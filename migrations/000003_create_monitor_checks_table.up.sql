CREATE TABLE IF NOT EXISTS monitor_checks (
    id SERIAL PRIMARY KEY,
    monitor_id INTEGER NOT NULL REFERENCES url_monitors(id) ON DELETE CASCADE,
    status_code INTEGER,
    response_time FLOAT,
    is_up BOOLEAN,
    ttfb FLOAT,
    content_size INTEGER,
    request_type VARCHAR(50) DEFAULT 'HEAD',
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
