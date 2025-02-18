CREATE TABLE url_monitors(
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    frequency_minutes INTEGER NOT NULL,
    timeout_seconds INTEGER,
    last_checked TIMESTAMP,
    expected_status_code INTEGER,
    status VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);



CREATE TABLE url_stats (
    id SERIAL PRIMARY KEY,
    monitor_id INTEGER NOT NULL REFERENCES url_monitors(id) ON DELETE CASCADE,
    status_code INTEGER,
    response_time FLOAT,
    is_up BOOLEAN,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)