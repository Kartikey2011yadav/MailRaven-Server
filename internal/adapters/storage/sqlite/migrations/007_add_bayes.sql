-- Migration: 007_add_bayes.sql
-- Description: Creates tables for Bayesian spam filtering
-- Created: 2026-01-31

-- Global statistics for probability calculation
CREATE TABLE bayes_global (
    key TEXT PRIMARY KEY, -- 'spam_total', 'ham_total'
    value INTEGER NOT NULL DEFAULT 0
);

-- Initialize global counters
INSERT INTO bayes_global (key, value) VALUES ('spam_total', 0);
INSERT INTO bayes_global (key, value) VALUES ('ham_total', 0);

-- The token corpus
CREATE TABLE bayes_tokens (
    token TEXT PRIMARY KEY,
    spam_count INTEGER NOT NULL DEFAULT 0,
    ham_count INTEGER NOT NULL DEFAULT 0
) WITHOUT ROWID;

-- Note: No separate index needed for 'token' as it is PRIMARY KEY
