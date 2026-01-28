# Data Model: PostgreSQL Schema

This document defines the schema mapping for the new PostgreSQL adapter.

## Type Logic

| SQLite Type | PostgreSQL Type | Notes |
| :--- | :--- | :--- |
| `TEXT` | `TEXT` or `VARCHAR` | Use `TEXT` generally, `VARCHAR(N)` for strict limits (e.g. email) |
| `INTEGER` | `BIGINT` | Go `int64` maps to Postgres `BIGINT` |
| `DATETIME` | `TIMESTAMPTZ` | Always use timezone-aware timestamps |
| `BOOLEAN` | `BOOLEAN` | Native boolean support |
| `BLOB` | `BYTEA` | For binary data |

## Schema Definitions

### Users Table (`users`)
```sql
CREATE TABLE users (
    email TEXT PRIMARY KEY,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);
```

### Domains Table (`domains`)
```sql
CREATE TABLE domains (
    name TEXT PRIMARY KEY,
    dkim_private_key TEXT,
    dkim_selector TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Emails Table (`emails`)
```sql
CREATE TABLE emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_email TEXT NOT NULL REFERENCES users(email) ON DELETE CASCADE,
    subject TEXT,
     -- Postgres-specific: Use array for headers or JSONB?
     -- Decision: Use JSONB for distinct headers to allow indexing
    headers JSONB, 
    body TEXT,
    folder VARCHAR(50) DEFAULT 'Inbox',
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Index for folder listing
CREATE INDEX idx_emails_user_folder ON emails(user_email, folder);
```

### Queue Table (`queue`)
```sql
CREATE TABLE queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payload JSONB NOT NULL, -- Flexible payload storage
    status VARCHAR(20) NOT NULL,
    next_attempt_at TIMESTAMPTZ NOT NULL,
    attempts INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Index for finding next job
CREATE INDEX idx_queue_process ON queue(status, next_attempt_at);
```

## Migration Strategy
- Store migrations in `internal/adapters/storage/postgres/migrations`.
- Use `001_init.sql` naming convention.
- App startup will check for `schema_version` table and apply missing migrations (Parallels SQLite logic).
