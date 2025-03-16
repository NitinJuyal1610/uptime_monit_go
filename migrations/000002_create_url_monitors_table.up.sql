CREATE TABLE IF NOT EXISTS url_monitors(
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    frequency_minutes INTEGER NOT NULL,
    timeout_seconds INTEGER,
    last_checked TIMESTAMP,
    expected_status_code INTEGER,
    status VARCHAR(50),
    max_fail_threshold INTEGER,
    consecutive_fails INTEGER DEFAULT 0,
    alert_email VARCHAR(50),
    collect_detailed_data BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
