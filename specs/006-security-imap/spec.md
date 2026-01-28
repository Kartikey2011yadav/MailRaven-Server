# Feature Specification: Security & IMAP Groundwork

**Feature Branch**: `006-security-imap`
**Created**: 2026-01-28
**Status**: Draft
**Input**: User description: "Implementing Spam Filtering (Rspamd integration or internal Bayesian) and considering IMAP4 groundwork."

## User Scenarios & Testing

### User Story 1 - Spam Filtering (Priority: P1)

As an administrator, I want incoming emails to be scanned for spam so that my users receive fewer unwanted messages.

**Why this priority**: Core security requirement for a production mail server.

**Independent Test**: Send an email containing the "GTUBE" test string; verify it is rejected or flagged.

**Acceptance Scenarios**:

1. **Given** valid configuration, **When** receiving an email with GTUBE string, **Then** the server rejects it with a 5xx error (or adds X-Spam header).
2. **Given** a normal email, **When** receiving, **Then** the server accepts it and adds an `Authentication-Results` header.
3. **Given** Rspamd is configured, **When** receiving an email, **Then** the system sends headers/body to Rspamd HTTP API and actions based on the response.

### User Story 2 - DNSBL Protection (Priority: P2)

As an administrator, I want to reject connections from known spam IP addresses using DNSBLs (like Zen.spamhaus.org) to reduce load.

**Why this priority**: Efficient first line of defense before processing data.

**Independent Test**: Configure a local DNSBL mock or use a test IP (127.0.0.2) against standard lists.

**Acceptance Scenarios**:

1. **Given** a sender IP listed in the configured DNSBL, **When** connecting, **Then** the server rejects the connection (or MAIL FROM) with a 5xx error.
2. **Given** a clean IP, **When** connecting, **Then** the server proceeds normally.
3. **Given** multiple DNSBLs configured, **When** connecting, **Then** the server checks them internally (sequential or parallel).

### User Story 3 - IMAP Connectivity (Priority: P2)

As a standard email client (e.g., Thunderbird), I want to connect to the server via IMAP ports (143/993) so that I can verify the server is reachable.

**Why this priority**: First step towards supporting standard clients.

**Independent Test**: `telnet localhost 143` returns a valid IMAP greeting.

**Acceptance Scenarios**:

1. **Given** the server is running, **When** connecting to port 143, **Then** a `* OK [CAPABILITY ...] MailRaven Ready` greeting is received.
2. **Given** the server is running, **When** connecting to port 993, **Then** a TLS handshake occurs before the greeting.
3. **Given** an open connection, **When** sending `A01 CAPABILITY`, **Then** the server lists supported capabilities (IMAP4rev1, STARTTLS, AUTH=PLAIN).

### User Story 4 - IMAP Authentication (Priority: P2)

As a user, I want to authenticate via IMAP so that I can establish a session.

**Why this priority**: Essential for identifying the user before any mailbox operations.

**Independent Test**: Send `A02 LOGIN user pass` and receive `OK`.

**Acceptance Scenarios**:

1. **Given** valid credentials, **When** sending `LOGIN user@domain.com password`, **Then** the server returns `A02 OK Logged in`.
2. **Given** invalid credentials, **When** sending `LOGIN`, **Then** the server returns `A02 NO Authentication failed`.
3. **Given** an unencrypted connection on port 143, **When** sending `LOGIN`, **Then** the server returns `BAD` (require STARTTLS first) unless configured to allow insecure.

## Functional Requirements

### 1. Spam Protection (Rspamd & DNSBL)
*   **DNSBL Hook**: Integrate into SMTP `Connect` or `MAIL FROM` phase.
    *   Check IP against configured lists (`zen.spamhaus.org`, `bl.spamcop.net`).
    *   Cache results (optional/future) or query live.
*   **Rspamd Hook**: Integrate into SMTP `DATA` phase.
    *   Send headers/body to Rspamd HTTP API (`/checkv2`).
    *   Handle response:
        *   `Action: reject` -> SMTP 550.
        *   `Action: add header` -> Add `X-Spam-Status: Yes`.
        *   `Action: soft reject` -> SMTP 451.
*   **Configuration**:
    *   `spam.dnsbls` (list of strings)
    *   `spam.rspamd_url` (string)
    *   `spam.enabled` (bool)

### 2. IMAP Listener
*   **Ports**: 143 (TCP+STARTTLS), 993 (TCP+TLS).
*   **Concurrency**: Handle multiple concurrent connections (goroutines).
*   **State Machine**: Implement `Not Authenticated` -> `Authenticated` states.
*   **Timeout**: Idle connections should timeout after 30 minutes (RFC standard).

### 3. IMAP Commands (Basics)
*   `CAPABILITY`: Return supported features.
*   `NOOP`: Keepalive.
*   `LOGOUT`: Close connection.
*   `STARTTLS`: Upgrade connection.
*   `LOGIN`: Plaintext auth.
*   `AUTHENTICATE PLAIN`: SASL auth.

### 4. RFC Compliance & Libraries
*   **RFC Support**: Ensure robust parsing for RFC 5322 (Message Format) and RFC 3501 (IMAP).
*   **Validation**: Enhance existing SPF/DKIM/DMARC validators (in `internal/adapters/smtp/validators`) to support strict RFC compliance if needed for Rspamd integration.
*   **Missing Packages**:
    *   `rfc`: General purpose RFC parsing (e.g., address parsing, date parsing).
    *   `iprev`: Reverse IP lookup logic (needed for Rspamd/DNSBL).
    *   `tlsrpt`: future scope, but keep in mind for architecture.

## Success Criteria

1.  **Spam Detection**: 100% of GTUBE emails are identified as spam.
2.  **IMAP Compliance**: Passes specific subset of IMAP compliance tests (Connection, Auth, Capability) without syntax errors.
3.  **Performance**: Spam check adds less than 500ms latency to SMTP transaction.
4.  **Security**: IMAP forces TLS before Auth (unless strictly configured otherwise).

## Assumptions

*   We are NOT implementing mailbox syncing (SELECT, FETCH, SEARCH) in this feature.
*   We reuse the existing `UserRepository` for IMAP authentication.
*   Certificates used for SMTP/HTTP are reused for IMAP.

## Key Entities

*   **ImapSession**: Represents a connected client state.
*   **SpamScorer**: Interface for the spam engine.
