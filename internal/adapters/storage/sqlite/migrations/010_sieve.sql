-- Migration: 010_sieve.sql
-- Description: Create tables for Sieve Filtering (Scripts and Vacation)

CREATE TABLE IF NOT EXISTS sieve_scripts (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    content TEXT NOT NULL,
    is_active INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

-- Ensure only one active script per user is enforced by application logic (transactional SetActive),
-- but we index it for speed.
CREATE INDEX IF NOT EXISTS idx_sieve_scripts_active ON sieve_scripts(user_id) WHERE is_active = 1;

CREATE TABLE IF NOT EXISTS vacation_trackers (
    user_id TEXT NOT NULL,
    sender_email TEXT NOT NULL,
    last_sent_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, sender_email)
);
