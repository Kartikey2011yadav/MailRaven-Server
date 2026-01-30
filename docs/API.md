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

### Autodiscover (Configuration)
- `POST /autodiscover/autodiscover.xml`: Microsoft Outlook autoconfig protocol.
- `GET /mail/config-v1.1.xml`: Mozilla Thunderbird autoconfig protocol.
- Note: These endpoints return XML and do not require JWT authentication (public).

### Messages
- `GET /messages`: List messages. Supports pagination (cursor-based) and filtering (folder, read status).
- `GET /messages/{id}`: Get full message details.
- `GET /messages/{id}/raw`: Get the raw MIME source.
- `POST /messages/send`: Submit an email for delivery.
- `PATCH /messages/{id}`: Update specific fields (e.g., mark as read/archived).

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
