# Data Model: Sieve Filtering

## Entities

### SieveScripts
Stores the Sieve scripts uploaded by users via ManageSieve or Web Admin.

```sql
CREATE TABLE sieve_scripts (
    id TEXT PRIMARY KEY,       -- UUID
    user_id TEXT NOT NULL,     -- References users(id)
    name TEXT NOT NULL,        -- Script name (from ManageSieve PUTSCRIPT)
    content TEXT NOT NULL,     -- The Sieve script body
    is_active INTEGER DEFAULT 0, -- Boolean (0/1), strict constraint: only 1 active per user
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

-- Index for fast retrieval during delivery
CREATE INDEX idx_sieve_scripts_active ON sieve_scripts(user_id) WHERE is_active = 1;
```

### VacationTrackers
Tracks auto-replies to prevent loops and storming. RFC 5230 compliant.

```sql
CREATE TABLE vacation_trackers (
    user_id TEXT NOT NULL,          -- Who is on vacation
    sender_email TEXT NOT NULL,     -- Who we replied to
    last_sent_at TIMESTAMP NOT NULL,-- When we replied
    PRIMARY KEY (user_id, sender_email)
);
```

## Relationships
- `User` 1:N `SieveScripts`
- `User` 1:N `VacationTrackers`
