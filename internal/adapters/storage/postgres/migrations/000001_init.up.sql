-- Users table
CREATE TABLE IF NOT EXISTS users (
    email TEXT PRIMARY KEY,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Domains table
CREATE TABLE IF NOT EXISTS domains (
    name TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active BOOLEAN NOT NULL DEFAULT true,
    dkim_selector TEXT,
    dkim_private_key TEXT,
    dkim_public_key TEXT
);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,
    subject TEXT,
    snippet TEXT,
    body_path TEXT NOT NULL,
    read_state BOOLEAN NOT NULL DEFAULT false,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    spf_result TEXT,
    dkim_result TEXT,
    dmarc_result TEXT,
    dmarc_policy TEXT,
    -- Index for faster lookup by user (recipient) and date
    CONSTRAINT idx_messages_recipient_date UNIQUE (recipient, received_at, id) 
    -- Constraint is wrong? No, we likely want index. UNIQUE might be too strict if two emails arrive EXACT same time. 
);

CREATE INDEX IF NOT EXISTS idx_messages_recipient_received_at ON messages (recipient, received_at DESC);

-- Queue table
CREATE TABLE IF NOT EXISTS queue (
    id TEXT PRIMARY KEY,
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,
    blob_key TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    next_retry_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT
);

-- Index for queue polling
CREATE INDEX IF NOT EXISTS idx_queue_poll ON queue (status, next_retry_at ASC);

-- Search Index
CREATE TABLE IF NOT EXISTS messages_search (
    message_id TEXT PRIMARY KEY REFERENCES messages(id) ON DELETE CASCADE,
    tsv tsvector
);
CREATE INDEX IF NOT EXISTS idx_messages_search_tsv ON messages_search USING GIN(tsv);
