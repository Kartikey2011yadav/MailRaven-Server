# Feature Specification: Standard Client Compliance

**Feature Branch**: `007-standard-client-compliance`
**Created**: 2026-01-30
**Status**: Draft
**Input**: User description: "Implement standard client compliance (IMAP Core and Autodiscover)"

## User Scenarios & Testing

### User Story 1 - Instant Account Setup (Autodiscover) (Priority: P1)

As a user, I want my email client (Thunderbird, Outlook, Mobile Mail) to automatically configure server settings so that I don't have to manually enter hostnames and ports.

**Why this priority**: Essential for "It just works" user experience and reducing support friction.

**Independent Test**: Enter email/password in a fresh Thunderbird instance; it should find settings automatically.

**Acceptance Scenarios**:

1. **Given** a new Outlook/Thunderbird client, **When** user enters `user@domain.com` and password, **Then** the client automatically discovers IMAP/SMTP host `mail.domain.com` and correct ports/TLS settings.
2. **Given** a request to `/.well-known/autoconfig/mail/config-v1.1.xml`, **Then** the server returns a valid XML configuration for Mozilla clients.
3. **Given** a POST request to `/autodiscover/autodiscover.xml`, **Then** the server returns a valid XML configuration for Microsoft clients.

---

### User Story 2 - Basic Mailbox Usage (IMAP Core) (Priority: P1)

As a user, I want to view my folders and read my emails in my preferred client so that I can use MailRaven with my existing tools.

**Why this priority**: Core functionality. Without this, the IMAP service is useless.

**Independent Test**: Connect via a standard client and verify Inbox loads and emails are readable.

**Acceptance Scenarios**:

1. **Given** an authenticated IMAP session, **When** client sends `LIST "" *`, **Then** server returns list of folders (INBOX, Sent, Trash, etc.).
2. **Given** a selected mailbox, **When** client sends `FETCH 1:* (BODY[])`, **Then** server returns full message content for all emails.
3. **Given** a selected mailbox, **When** client sends `UID FETCH`, **Then** server returns persistent UIDs for synchronization.
4. **Given** a selected mailbox, **When** client sends `SELECT INBOX`, **Then** server returns `OK` with mailbox status (exists, recent, uidvalidity).

---

### User Story 3 - Instant Notifications (IMAP IDLE) (Priority: P2)

As a mobile user, I want to be notified immediately when new mail arrives so that I don't have to manually refresh the app.

**Why this priority**: Critical for modern mobile experience and battery life (avoids polling).

**Independent Test**: Connect via `openssl s_client`, send `IDLE`, then inject an email via SMTP. The client should see `* EXISTS` immediately.

**Acceptance Scenarios**:

1. **Given** a client in IDLE state, **When** a new email is delivered to the mailbox, **Then** server sends `* <N> EXISTS` untagged response immediately.
2. **Given** a client in IDLE state, **When** client sends `DONE`, **Then** server returns to command mode.

---

## Functional Requirements

### 1. IMAP Protocol Support (RFC 3501)
The IMAP server must implement the following commands currently missing or incomplete:

*   **Mailbox Operations**:
    *   `SELECT` / `EXAMINE`: Open mailbox, return `FLAGS`, `EXISTS`, `RECENT`, `UIDVALIDITY`.
    *   `LIST` / `LSUB`: List available mailboxes (Hierarchical support).
    *   `CREATE` / `DELETE` / `RENAME`: Manage mailboxes.
    *   `STATUS`: Return mailbox metrics (MESSAGES, UNSEEN).
    
*   **Message Operations**:
    *   `FETCH`: Support macros (`ALL`, `FAST`, `FULL`) and specific data items (`BODY`, `BODYSTRUCTURE`, `ENVELOPE`, `FLAGS`, `INTERNALDATE`, `RFC822.SIZE`, `UID`).
    *   `UID`: Support UID variants of commands (`UID FETCH`, `UID COPY`, `UID STORE`).
    *   `STORE`: Update flags (Seen, Answered, Flagged, Deleted).

*   **Extensions**:
    *   `IDLE` (RFC 2177): Real-time updates.

### 2. Autodiscover Service
The HTTP server must expose endpoints to support client auto-configuration:

*   **Mozilla Autoconfig**:
    *   Endpoint: `GET /.well-known/autoconfig/mail/config-v1.1.xml`
    *   Logic: Return XML with `incomingServer` (IMAP) and `outgoingServer` (SMTP) details based on the request domain or default config.

*   **Microsoft Autodiscover**:
    *   Endpoint: `POST /autodiscover/autodiscover.xml`
    *   Logic: Parse XML request, return XML response (POX) configured for IMAP/SMTP.

## Success Criteria

1.  **Client Compatibility**:
    *   **Thunderbird**: Auto-configures + Reads Email + IDLE works.
    *   **Outlook (New)**: Auto-configures + Reads Email.
    *   **iOS Mail**: Reads Email.
2.  **Compliance**:
    *   Passes `TestIMAP_Compliance` suite (all checks green).
3.  **Performance**:
    *   `IDLE` notifications delivered within < 5 seconds of message arrival.
    *   `FETCH` of 50 message headers takes < 1 second on local LAN.

## Assumptions

*   We are targeting `IMAP4rev1`.
*   Storage layer supports retrieving emails by `UID` range.
*   DNS SRV record configuration is out of scope for the *code* (it's documentation/deployment), but the server endpoints are in scope.

## Key Entities & Data Transformed

*   **IMAP Session**: Needs state machine expansion (Selected Mailbox context).
*   **Mailbox**: Representation of folder with counts/status.
*   **Autodiscover Config**: Structs for XML marshalling.
