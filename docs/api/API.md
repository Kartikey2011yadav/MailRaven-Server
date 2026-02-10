# MailRaven API Documentation

MailRaven exposes a RESTful JSON API. This API is the primary way for clients (mobile apps, web frontends) to interact with the mailbox.

## Base URL

By default: `http://localhost:8080/api/v1`

## Authentication

Authentication is handled via **JWT (JSON Web Tokens)**.
Clients must obtain a token via the login endpoint and include it in the `Authorization` header:

```
Authorization: Bearer <token>
```

## Core Endpoints

*(Note: Implemented routes can be found in `internal/adapters/http/server.go`)*

### Authentication
- `POST /auth/login`: Exchange credentials for a JWT.
- `POST /auth/refresh`: Refresh an expiring token.

### Autodiscover & Public Well-Known
- `POST /autodiscover/autodiscover.xml`: Microsoft Outlook autoconfig protocol.
- `GET /mail/config-v1.1.xml`: Mozilla Thunderbird autoconfig protocol.
- `GET /.well-known/mta-sts.txt`: MTA-STS policy file (served on `mta-sts.<domain>`).
- `POST /.well-known/tlsrpt`: TLS Reporting endpoint (JSON ingestion).
- Note: These endpoints are public and do not require JWT authentication.

### Messages
- `GET /messages`: List messages. Supports pagination (limit, offset) and filtering:
  - `mailbox`: Filter by folder (e.g., INBOX, Archive, Junk, Trash)
  - `is_read`: Filter by read status (true/false)
  - `is_starred`: Filter by starred status (true/false)
  - `start_date`, `end_date`: Filter by received date range.
- `GET /messages/{id}`: Get full message details.
- `GET /messages/{id}/raw`: Get the raw MIME source.
- `POST /messages/send`: Submit an email for delivery.
- `PATCH /messages/{id}`: Update message state.
  - `is_read`: Boolean
  - `is_starred`: Boolean
  - `mailbox`: Target mailbox name (Archive, Junk, Trash, INBOX)
- `POST /messages/{id}/spam`: Report message as Spam (moves to Junk + trains filter).
- `POST /messages/{id}/ham`: Report message as Ham (moves to Inbox + trains filter).

### Sieve Scripts
- `GET /sieve/scripts`: List all Sieve scripts for the authenticated user.
- `POST /sieve/scripts`: Upload a new Sieve script.
- `GET /sieve/scripts/{name}`: Download the content of a specific script.
- `DELETE /sieve/scripts/{name}`: Delete a script.
- `PUT /sieve/scripts/{name}/active`: Activate a script (deactivates others).

### Search
- `GET /search`: dedicated full-text search endpoint utilizing FTS5 or Postgres TSVECTOR.
  - Query parameters: `q` (search term), `from`, `has_attachment`.

### Management (Web Admin)
- `GET /admin/users`: List users.
- `POST /admin/users`: Create user.
- `GET /admin/domains`: List domains.
- `POST /admin/domains`: Add domain.
- `GET /admin/stats/overview`: Get system statistics.

### Monitoring
- `GET /metrics`: Prometheus formatted metrics (System health, Queue depth, Inbound/Outbound counts).

## Differences from standard IMAP

| Feature | IMAP | MailRaven API |
|---------|------|---------------|
| Protocol | TCP (Stateful) | HTTP (Stateless) |
| Format | MIME / Text | JSON |
| Sync | Complex (IDLE, UID) | Simple (Delta Sync tokens) |
| Search | Server-side (slow) | SQL-backed FTS (fast) |

## Implementation Details

The API layer translates JSON requests into Domain objects and passes them to the `Core` services. It does *not* contain business logic.
- Input validation handles structure (required fields, types).
- Core services handle business rules (permissions, state transitions).
