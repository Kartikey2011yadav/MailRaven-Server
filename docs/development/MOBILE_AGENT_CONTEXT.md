# Mobile Agent Context

**Purpose**: This document serves as the Source of Truth for the Mobile App AI Agent.
**Updated**: 2026-02-02

## 1. API & Authentication

### Base URL
-   Development: `http://localhost:1080/api` (WebAPI) or `http://localhost:8080` (Mox Web)
-   Production: `https://mail.yourdomain.com/api`

### Authentication
-   **Type**: Cookie-based or Token-based (TBD during implementation).
-   **Endpoint**: `POST /auth/login` (Standard) or `POST /api/v1/session`
-   **Credentials**: Email and Password.

## 2. Protocol Interactions

### IMAP (Primary Data Sync)
-   **Port**: 993 (SSL), 143 (StartTLS)
-   **Library Recommendation**: `go-imap` (if Go), or platform native.
-   **Optimization**: Use `CONDSTORE` (RFC 7162) for delta sync.
-   **Push**: Use `IDLE` (RFC 2177) for real-time updates.

### SMTP (Sending)
-   **Port**: 587 (Submission) or 465 (SSL).
-   **Auth**: PLAIN or LOGIN.

## 3. Data Models (Client Side)

-   **Account**: Email, Host, Port, Credentials (secure storage).
-   **Mailbox**: Name, UnseenCount, UIDValidity, HighestModSeq.
-   **Message**: UID, Flags, InternalDate, Envelope, BodyStructure.

## 4. Key Constraints

-   **Certificates**: Must allow "Accept Invalid Certs" in Dev mode.
-   **Background**: Sync must handle OS background execution limits.
-   **Offline**: All actions (read, move, delete) must correspond to an offline-queue.
