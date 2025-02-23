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



CREATE TABLE monitor_checks (
    id SERIAL PRIMARY KEY,
    monitor_id INTEGER NOT NULL REFERENCES url_monitors(id) ON DELETE CASCADE,
    status_code INTEGER,
    response_time FLOAT,
    is_up BOOLEAN,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)

CREATE INDEX idx_monitor_checks_monitor ON monitor_checks (monitor_id, is_up);
CREATE INDEX idx_monitor_checks_response_time ON monitor_checks (monitor_id, response_time);
