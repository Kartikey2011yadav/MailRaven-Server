-- Migration: 006_add_greylist.sql
-- Description: Creates table for SMTP greylisting
-- Created: 2026-01-31

CREATE TABLE greylist (
    ip_net TEXT NOT NULL,       -- Normalized IP (e.g. 192.168.1.0/24)
    sender TEXT NOT NULL,       -- Envelope MAIL FROM
    recipient TEXT NOT NULL,    -- Envelope RCPT TO
    first_seen_unix INTEGER NOT NULL,
    last_seen_unix INTEGER NOT NULL,
    blocked_count INTEGER DEFAULT 0,
    PRIMARY KEY (ip_net, sender, recipient)
) WITHOUT ROWID;

-- Index for efficient pruning of expired records
CREATE INDEX idx_greylist_last_seen ON greylist(last_seen_unix);
