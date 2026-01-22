# Feature Specification: Mobile Email Server

**Feature Branch**: `001-mobile-email-server`  
**Created**: 2026-01-22  
**Status**: Draft  
**Input**: User description: "Build a modern, mobile-centric email server inspired by 'Mox'. The server must: 1. Core: Listen on ports 25 (SMTP) and 443 (HTTPS). 2. Inbound: Accept emails, parse MIME parts, and validate sender reputation (SPF/DMARC) using logic adapted from the local 'mox' repository. 3. Storage: Store email metadata (Sender, Subject, Snippet, ReadState) in a local SQLite database for instant querying by the API. Store full raw bodies in a compressed file store. 4. Client API: Instead of IMAP, expose a secure JSON REST API for the 'MailRaven' Android app to sync emails. 5. Administration: Replicate Mox's 'Quickstart' feature—on first run, generate config and print required DNS records (MX, SPF, DKIM) to the console."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Initial Server Setup and Configuration (Priority: P1)

Administrator sets up MailRaven server for the first time by running a quickstart command. The system generates all necessary configuration and displays DNS records that must be created for the domain to receive email.

**Why this priority**: Without proper setup, no emails can be received. This is the foundation for all other functionality.

**Independent Test**: Run `mailraven quickstart admin@example.com` and verify: (1) config files are created, (2) DNS records (MX, SPF, DKIM) are printed to console, (3) server can start successfully.

**Acceptance Scenarios**:

1. **Given** fresh installation with no config files, **When** administrator runs `mailraven quickstart admin@mydomain.com`, **Then** system generates config files and prints required DNS records (MX, SPF, DKIM, DMARC) to console
2. **Given** quickstart has been run, **When** administrator adds DNS records and starts server, **Then** server listens on port 25 (SMTP) and port 443 (HTTPS) without errors
3. **Given** server is running, **When** administrator sends test email to configured address, **Then** email is accepted and stored successfully

---

### User Story 2 - Receive and Store Emails (Priority: P1)

External mail servers send emails to MailRaven on port 25. The system accepts emails, validates sender reputation using SPF/DMARC checks, parses MIME structure, and stores messages durably.

**Why this priority**: Core email reception is the primary purpose. Without this, the server cannot function as a mail server.

**Independent Test**: Send email via external SMTP client to the server and verify: (1) SMTP responds with "250 OK", (2) message appears in SQLite database with metadata, (3) full message body exists in file store, (4) SPF/DMARC results are recorded.

**Acceptance Scenarios**:

1. **Given** server is running and DNS is configured, **When** external server sends email to user@mydomain.com, **Then** server performs SPF lookup, DMARC validation, accepts email with "250 OK", and stores message atomically (SQLite + file)
2. **Given** email with multiple MIME parts (text, HTML, attachments), **When** server receives email, **Then** system parses all parts, extracts metadata (sender, subject, snippet), and stores raw message compressed in file store
3. **Given** email from sender with invalid SPF record, **When** server receives email, **Then** SPF result is marked as "fail", DMARC policy is evaluated, and decision (accept/reject) follows policy
4. **Given** server receives email during SMTP DATA phase, **When** disk write or fsync fails, **Then** server responds with temporary failure (4xx) instead of "250 OK"

---

### User Story 3 - Mobile App Syncs Emails via API (Priority: P1)

MailRaven Android app connects to the server's JSON REST API over HTTPS to retrieve email list, mark messages as read, and download full message content.

**Why this priority**: The mobile client is the primary interface for users to access their email. This is what differentiates MailRaven from traditional mail servers.

**Independent Test**: Use Android app (or API client) to authenticate, fetch email list, mark message as read, and verify changes persist. All operations use JSON over HTTPS.

**Acceptance Scenarios**:

1. **Given** user has received 10 emails, **When** mobile app authenticates and calls GET /api/v1/messages?limit=20, **Then** API returns JSON array with email metadata (id, sender, subject, snippet, read_state, timestamp) for all 10 messages
2. **Given** email list is displayed in app, **When** user taps on unread message, **Then** app calls GET /api/v1/messages/{id} and receives full message content including all MIME parts
3. **Given** user has opened a message, **When** app calls PATCH /api/v1/messages/{id} with read_state=true, **Then** message is marked as read in SQLite and subsequent API calls reflect updated state
4. **Given** user has 1000+ emails in mailbox, **When** app requests messages with pagination (limit=50, offset=0), **Then** API returns first 50 messages quickly (<100ms) without loading entire dataset

---

### User Story 4 - Authentication and Authorization (Priority: P2)

Mobile app and administrators authenticate to the server using secure credentials. API endpoints require valid authentication tokens to prevent unauthorized access.

**Why this priority**: Security is critical, but basic email reception can work without app authentication for initial testing. This is needed before production deployment.

**Independent Test**: Attempt API requests without auth token (receive 401), authenticate with valid credentials (receive token), use token for subsequent requests (succeed).

**Acceptance Scenarios**:

1. **Given** mobile app starts, **When** user enters email and password and taps login, **Then** app sends POST /api/v1/auth/login with credentials and receives JWT token
2. **Given** app has valid auth token, **When** app calls any /api/v1/* endpoint with Authorization header, **Then** server validates token and processes request
3. **Given** app calls API without auth token, **When** server receives request, **Then** server responds with 401 Unauthorized
4. **Given** auth token has expired (1 week old), **When** app uses expired token, **Then** server responds with 401 and app prompts user to re-authenticate

---

### User Story 5 - Send Outbound Email via API (Priority: P3)

Mobile app allows users to compose and send emails. App submits message via API endpoint, server generates DKIM signature, and delivers email to recipient's mail server via SMTP.

**Why this priority**: Receiving email is more critical than sending initially. Sending can be added after core reception and API functionality are stable.

**Independent Test**: Use app to compose email, submit via POST /api/v1/messages/send, verify server attempts SMTP delivery to recipient's mail server and message is stored in sent folder.

**Acceptance Scenarios**:

1. **Given** user composes email in app, **When** app posts JSON to /api/v1/messages/send with to, subject, body fields, **Then** server generates DKIM signature and queues message for delivery
2. **Given** outbound message is queued, **When** server processes delivery queue, **Then** server performs MX lookup for recipient domain and delivers via SMTP to recipient's mail server
3. **Given** delivery succeeds, **When** recipient's server accepts with "250 OK", **Then** server records delivery status and message appears in user's sent folder via API
4. **Given** delivery fails (recipient server unreachable), **When** server attempts delivery, **Then** server retries with exponential backoff and reports delivery status via API

---

### Edge Cases

- What happens when SQLite database is corrupted? (Server should detect on startup with integrity_check and fail gracefully with clear error message)
- How does system handle emails larger than available disk space? (Reject with SMTP 4xx error before accepting with "250 OK")
- What if DKIM/SPF DNS records are missing during quickstart? (Generate records anyway and warn user they must be added)
- How does API handle concurrent requests marking same message read? (Use optimistic locking or last-write-wins with idempotency)
- What happens when file store and SQLite get out of sync? (Reconciliation tool to detect orphaned files or missing bodies)
- How does server handle malformed MIME structures? (Parse strictly per RFC, reject with SMTP error if structure is invalid)
- What if mobile app requests 1M messages at once? (API enforces max limit=1000 per request, returns 400 Bad Request if exceeded)

## Requirements *(mandatory)*

### Functional Requirements

**Core Server Infrastructure:**

- **FR-001**: System MUST listen on port 25 (SMTP) for inbound email from external mail servers
- **FR-002**: System MUST listen on port 443 (HTTPS) for secure API connections from mobile clients
- **FR-003**: System MUST run on Linux servers with Go runtime (CGO-free deployment)

**Email Reception & Validation:**

- **FR-004**: System MUST accept SMTP connections and implement RFC 5321 SMTP protocol
- **FR-005**: System MUST validate sender reputation using SPF (RFC 7208) checks by performing DNS lookups
- **FR-006**: System MUST validate DMARC (RFC 7489) policies for received emails
- **FR-007**: System MUST parse MIME structure (RFC 2045) of received emails including multipart messages
- **FR-008**: System MUST extract text and HTML body parts from MIME messages for snippet generation
- **FR-009**: System MUST reply "250 OK" to SMTP DATA command ONLY after message is durably stored (SQLite + file + fsync)

**Storage & Durability:**

- **FR-010**: System MUST store email metadata in SQLite database with fields: message_id, sender, recipient, subject, snippet (first 200 chars), read_state, timestamp, spf_result, dmarc_result
- **FR-011**: System MUST store full raw email message body in compressed file store (gzip compression)
- **FR-012**: System MUST use atomic transactions with PRAGMA synchronous=FULL for SQLite writes
- **FR-013**: System MUST call fsync() on both SQLite database and message files before SMTP acknowledgment
- **FR-014**: System MUST verify SQLite database integrity on startup using PRAGMA integrity_check
- **FR-015**: System MUST reconcile SQLite metadata with file store on startup and log any inconsistencies

**Mobile REST API:**

- **FR-016**: System MUST expose JSON REST API on /api/v1/* endpoints over HTTPS
- **FR-017**: API MUST support GET /api/v1/messages endpoint with pagination (limit, offset) returning message metadata array
- **FR-018**: API MUST support GET /api/v1/messages/{id} endpoint returning full message content
- **FR-019**: API MUST support PATCH /api/v1/messages/{id} endpoint to update read_state
- **FR-020**: API MUST support POST /api/v1/auth/login endpoint for authentication returning JWT token
- **FR-021**: API MUST validate JWT tokens on all /api/v1/* endpoints except /auth/login
- **FR-022**: API responses MUST support gzip compression when client sends Accept-Encoding: gzip header
- **FR-023**: API MUST enforce maximum page size of 1000 messages per request

**Administration & Configuration:**

- **FR-024**: System MUST provide quickstart command that accepts email address and generates initial configuration
- **FR-025**: Quickstart MUST generate DKIM private key and public key for DNS TXT record
- **FR-026**: Quickstart MUST print to console the DNS records required: MX record, SPF TXT record, DKIM TXT record, DMARC TXT record
- **FR-027**: System MUST load configuration from file on startup specifying domain, listen ports, storage paths, DKIM key path

**Security:**

- **FR-028**: System MUST use TLS for HTTPS API endpoints with valid certificates
- **FR-029**: System MUST support STARTTLS for SMTP connections (optional but recommended)
- **FR-030**: System MUST hash and salt passwords for authentication (use bcrypt or argon2)
- **FR-031**: System MUST implement rate limiting on API endpoints (100 requests/minute per IP)
- **FR-032**: System MUST log all authentication attempts (success and failure) with IP address and timestamp

### Key Entities

- **Message**: Represents an email message. Attributes: unique ID, sender address, recipient address, subject line, message snippet (first 200 chars of body), read state (boolean), received timestamp, SPF validation result, DMARC validation result, file store path.

- **MessageBody**: Full raw email message stored in compressed file format. Attributes: message ID (references Message), file path, compressed size, original size, compression algorithm.

- **User**: Represents a mailbox user with credentials. Attributes: email address, password hash, created timestamp, last login timestamp.

- **AuthToken**: JWT token for API authentication. Attributes: token string, user email, expiration timestamp, issued timestamp.

- **SMTPSession**: Temporary state for active SMTP connection. Attributes: session ID, remote IP, sender (MAIL FROM), recipients (RCPT TO list), connection timestamp, bytes received.

- **DNSRecord**: Generated DNS configuration for domain. Attributes: record type (MX/SPF/DKIM/DMARC), record name, record value, TTL.

## Success Criteria *(mandatory)*

### Measurable Outcomes

**Core Functionality:**

- **SC-001**: Administrator can run quickstart command and have server receiving emails within 10 minutes (after DNS propagation)
- **SC-002**: Server successfully receives and stores 100 test emails from external SMTP servers with 100% delivery success rate
- **SC-003**: All received emails pass "250 OK" durability test - kill -9 server immediately after acknowledgment and verify 100% message recovery on restart

**Protocol Compliance:**

- **SC-004**: Server correctly validates SPF for test emails from Gmail, Outlook, Yahoo with 100% accuracy matching mox's validation results
- **SC-005**: Server correctly parses and stores MIME emails with 10+ parts (text, HTML, attachments) without data loss

**Mobile API Performance:**

- **SC-006**: Mobile app can authenticate and fetch list of 50 emails in under 500ms (P95 latency) over 4G connection
- **SC-007**: API pagination allows app to scroll through 10,000 messages without loading entire dataset (memory stays under 100MB)
- **SC-008**: API serves 100 concurrent mobile clients without response time degradation beyond 20%

**Reliability & Data Safety:**

- **SC-009**: Zero data loss after 1000 simulated crashes (kill -9) during various SMTP phases
- **SC-010**: SQLite database integrity check passes 100% after crash recovery tests
- **SC-011**: Server handles disk full scenario gracefully by rejecting new emails with SMTP 4xx (no "250 OK" followed by silent loss)

**Security:**

- **SC-012**: API rejects 100% of requests without valid JWT token (no unauthorized access)
- **SC-013**: Rate limiting prevents single IP from exceeding 100 requests/minute (tested with load generator)

**Developer Experience:**

- **SC-014**: Developer can set up local test environment and send test email to server in under 15 minutes following quickstart guide
- **SC-015**: DNS record generation by quickstart produces valid records that pass syntax validation (dig, nslookup)

## Assumptions

- **A-001**: MailRaven will initially support receiving email for a single domain (multi-domain support is future enhancement)
- **A-002**: Mobile app will be Android-only initially (iOS API compatibility considered but not required for MVP)
- **A-003**: Server will run on Linux (Ubuntu 20.04+) with public IP address and proper DNS configuration
- **A-004**: Administrator has basic understanding of DNS record management
- **A-005**: Initial deployment targets <1000 users with <50,000 total messages (scaling beyond this is future work)
- **A-006**: SPF/DMARC validation will use mox's implementation patterns but may not implement 100% of edge cases initially
- **A-007**: Email size limit is 25MB (standard industry practice)
- **A-008**: SQLite is sufficient for single-server deployment (distributed storage not required for MVP)
- **A-009**: DKIM signing for outbound email uses RSA-2048 keys (Ed25519 support is optional)
- **A-010**: API authentication uses JWT with 7-day expiration (refresh tokens are future enhancement)

## Out of Scope (Explicitly Excluded)

- **OS-001**: IMAP/POP3 protocols - MailRaven uses custom JSON API instead
- **OS-002**: Web-based email client (webmail) - mobile app is primary interface
- **OS-003**: Email forwarding rules and filters - simple storage only for MVP
- **OS-004**: Calendar, contacts, tasks integration - email-only focus
- **OS-005**: Multi-user admin panel - single admin via config file
- **OS-006**: Spam filtering and machine learning - rely on SPF/DMARC only for MVP
- **OS-007**: Backup/restore tooling - administrator uses standard file backup
- **OS-008**: High availability clustering - single server deployment
- **OS-009**: Email search with full-text indexing - simple metadata queries only
- **OS-010**: Attachment virus scanning - client-side handling
