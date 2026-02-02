# Mobile Agent Context

This document provides the context required for a Mobile Development AI Agent to build a mobile client for MailRaven independently. It defines the API contract, authentication flows, and data models.

## 1. Authentication

The API uses JWT (JSON Web Tokens) for authentication.

### **Login**
- **Endpoint**: `POST /api/v1/auth/login`
- **Content-Type**: `application/json`
- **Request**:
  ```json
  {
    "email": "user@example.com",
    "password": "your_password"
  }
  ```
- **Response (200 OK)**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsIn...",
    "expires_at": "2026-02-09T10:00:00Z"
  }
  ```
- **Usage**: Include the token in the `Authorization` header for all protected requests:
  ```
  Authorization: Bearer <token>
  ```

## 2. Mailbox Synchronization

The API supports both paginated fetching and delta synchronization (getting changes since a timestamp).

### **List Messages (Pagination)**
- **Endpoint**: `GET /api/v1/messages`
- **Params**:
  - `limit` (int, default 50): Number of messages to return.
  - `offset` (int, default 0): Number of messages to skip.
- **Response (200 OK)**:
  ```json
  [
    {
      "id": "uuid-string",
      "sender": "sender@remote.com",
      "subject": "Hello World",
      "snippet": "This is the first 200 chars...",
      "received_at": "2026-02-02T12:00:00Z",
      "read": false,
      "mailbox": "INBOX"
    }
  ]
  ```

### **Delta Sync (Updates)**
- **Endpoint**: `GET /api/v1/messages/since`
- **Params**:
  - `timestamp` (string, RFC3339): Fetch messages received after this time.
- **Example**: `/api/v1/messages/since?timestamp=2026-02-02T12:00:00Z`
- **Response**: Same as List Messages.

### **Get Message Details**
- **Endpoint**: `GET /api/v1/messages/{id}`
- **Response (200 OK)**:
  ```json
  {
      "id": "uuid-string",
      "from": "sender@remote.com",
      "to": ["user@example.com"],
      "subject": "Full Subject",
      "date": "2026-02-02T12:00:00Z",
      "body_text": "Plain text body...",
      "body_html": "<p>HTML body...</p>"
  }
  ```

## 3. Sending Email

- **Endpoint**: `POST /api/v1/messages/send`
- **Request**:
  ```json
  {
    "to": ["recipient@example.com", "cc@example.com"],
    "subject": "Subject Line",
    "body": "Message plain text content..."
  }
  ```
- **Response (200 OK)**:
  ```json
  {
    "status": "queued",
    "id": "queue-id"
  }
  ```

## 4. IMAP/SMTP Connection Details

For clients preferring standard protocols:

- **IMAP Server**:
  - Port: 143 (STARTTLS) or 993 (Implicit TLS)
  - Auth: PLAIN
- **SMTP Server**:
  - Port: 25 (STARTTLS) or 587 (Submission)
  - Auth: PLAIN (Required for relay)

## 5. Verification Commands

Use these curl commands to verify the local server:

```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login -d '{"email":"user@example.com","password":"password"}' | jq -r .token)

# 2. List Messages
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/messages

# 3. Send Email
curl -X POST -H "Authorization: Bearer $TOKEN" -d '{"to":["self@example.com"],"subject":"Test","body":"Content"}' http://localhost:8080/api/v1/messages/send
```
