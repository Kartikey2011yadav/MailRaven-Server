CREATE TABLE domains (
    name TEXT PRIMARY KEY,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    dkim_selector TEXT,
    dkim_private_key TEXT,
    dkim_public_key TEXT
);

-- Index for searching domains
CREATE INDEX idx_domains_active ON domains(active);
