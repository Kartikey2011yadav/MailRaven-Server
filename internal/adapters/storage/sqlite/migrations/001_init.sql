-- MailRaven Database Schema v1
-- SQLite with Write-Ahead Logging and full durability

-- Enable Write-Ahead Logging for better concurrent performance
PRAGMA journal_mode=WAL;

-- Full durability: fsync on every commit (required by constitution)
PRAGMA synchronous=FULL;

-- Foreign key enforcement
PRAGMA foreign_keys=ON;

-- =============================================================================
-- Users Table
-- =============================================================================

CREATE TABLE users (
    email TEXT PRIMARY KEY,
    password_hash TEXT NOT NULL,
    created_at INTEGER NOT NULL,  -- Unix timestamp
    last_login_at INTEGER NOT NULL
);

CREATE INDEX idx_users_last_login ON users(last_login_at DESC);

-- =============================================================================
-- Messages Table
-- =============================================================================

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,          -- Email Message-ID header
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,           -- References users(email)
    subject TEXT NOT NULL,
    snippet TEXT NOT NULL,
    body_path TEXT NOT NULL,           -- Path in blob store
    read_state INTEGER NOT NULL,       -- 0 = unread, 1 = read
    received_at INTEGER NOT NULL,      -- Unix timestamp
    
    spf_result TEXT NOT NULL,          -- "pass", "fail", "softfail", "neutral", "none"
    dkim_result TEXT NOT NULL,         -- "pass", "fail", "none"
    dmarc_result TEXT NOT NULL,        -- "pass", "fail", "none"
    dmarc_policy TEXT NOT NULL,        -- "none", "quarantine", "reject"
    
    FOREIGN KEY (recipient) REFERENCES users(email) ON DELETE CASCADE
);

-- Performance indexes
CREATE INDEX idx_messages_recipient ON messages(recipient, received_at DESC);

-- =============================================================================
-- Outbound Queue Table
-- =============================================================================

CREATE TABLE queue (
    id TEXT PRIMARY KEY,
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,
    blob_key TEXT NOT NULL,        -- Key in blob store containing signed message
    status TEXT NOT NULL,          -- PENDING, PROCESSING, SENT, FAILED, RETRYING
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    next_retry_at INTEGER NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT
);

-- Indexes for efficient queue processing
-- Find next message: status IN ('PENDING', 'RETRYING') AND next_retry_at <= NOW
CREATE INDEX idx_queue_process ON queue(status, next_retry_at ASC);

CREATE INDEX idx_messages_received_at ON messages(received_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender);
CREATE INDEX idx_messages_read_state ON messages(recipient, read_state);
CREATE UNIQUE INDEX idx_messages_message_id ON messages(message_id);

-- =============================================================================
-- Full-Text Search (FTS5)
-- =============================================================================

-- Virtual table for full-text search
CREATE VIRTUAL TABLE messages_fts USING fts5(
    message_id UNINDEXED,              -- Foreign key, not searchable
    recipient UNINDEXED,               -- Filter field, not searchable
    sender,                            -- Searchable
    subject,                           -- Searchable
    body_text,                         -- Searchable (first 10KB of plaintext)
    tokenize='porter unicode61'       -- Porter stemming + Unicode tokenization
);

-- Keep FTS in sync with messages table
-- Note: Triggers add slight overhead but ensure consistency
CREATE TRIGGER messages_fts_insert AFTER INSERT ON messages BEGIN
    INSERT INTO messages_fts(message_id, recipient, sender, subject, body_text)
    VALUES (NEW.id, NEW.recipient, NEW.sender, NEW.subject, '');
    -- Body text populated asynchronously by search indexer
END;

CREATE TRIGGER messages_fts_delete AFTER DELETE ON messages BEGIN
    DELETE FROM messages_fts WHERE message_id = OLD.id;
END;
