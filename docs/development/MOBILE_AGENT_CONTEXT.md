# Mobile Agent Context

## Overview

This document serves as the "Source of Truth" for the AI Agent building the MailRaven Mobile Client.
It abstracts the internal complexity of the server and exposes only the public API contracts,
authentication mechanism, and data models required for the client functioning.

## 1. Authentication

MailRaven uses JWT-based authentication.

### Login
`POST /api/v1/auth/login`

**Request**:
```json
{
  "email": "user@example.com",
  "password": "secret_password"
}
```

**Response**:
```json
{
  "token": "ey...",
  "expires_in": 3600
}
```

The returned token must be sent in the `Authorization` header as `Bearer <token>` for all subsequent requests.

## 2. Mailbox Synchronization

### Sync Strategy
MailRaven supports a "delta sync" mechanism using cursors/modification sequences (modseq).
(TODO: Verify implementation details in Phase 4)

### Endpoints
- `GET /api/v1/mailboxes`: List all mailboxes
- `GET /api/v1/mailboxes/{id}/messages`: Fetch messages (supports cursor-based pagination)

## 3. Sending Email

### Submit
`POST /api/v1/submission`

**Payload**:
MIME or JSON structured email.
(TODO: Define exact payload in Phase 4)

## 4. Constraints

- **TLS**: Use TLS 1.2+ for all connections in production.
- **Offline**: The client MUST cache data locally (SQLite/Realm) and queue actions when offline.
