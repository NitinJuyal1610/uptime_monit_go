CREATE INDEX IF NOT EXISTS idx_monitor_checks_monitor ON monitor_checks (monitor_id, is_up);
CREATE INDEX IF NOT EXISTS idx_monitor_checks_response_time ON monitor_checks (monitor_id, response_time);
