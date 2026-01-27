# Data Model: Web Admin & Domains

## Entities

### Domain
Represents an authorized internet domain for email services.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | The domain name (e.g., "example.com"). Primary Key. |
| `created_at` | `timestamp` | Yes | When the domain was added. |
| `is_primary` | `boolean` | No | Virtual field (not in DB). True if from config. |

**Validation Rules**:
- `name`: Must be a valid FQDN (Regex check).
- `name`: Unique constraint in DB.

## Schema Changes (SQLite)

```sql
-- migration: 003_add_domains.sql

CREATE TABLE domains (
    name TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Note: Existing 'users' table has 'email'. 
-- We do not strictly normalize users to link to domains via ID 
-- because email is the key. 
-- Validation happens at Application Layer (Create User -> Check Domain exists).
```

## State Transitions

- **Life Cycle**: Created -> Active -> Deleted.
- **Dependencies**: Cannot delete a Domain if Users exist with that email suffix (Application Rule).
