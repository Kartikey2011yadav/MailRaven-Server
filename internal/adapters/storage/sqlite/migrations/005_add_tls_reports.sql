-- Migration: Add tls_reports table
CREATE TABLE IF NOT EXISTS tls_reports (
    id TEXT PRIMARY KEY,
    report_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    total_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    raw_json TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tls_reports_date ON tls_reports(start_date);
CREATE INDEX IF NOT EXISTS idx_tls_reports_provider ON tls_reports(provider);
