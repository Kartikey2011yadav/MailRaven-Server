# Feature Specification: Sieve Filtering

**Feature Branch**: `010-sieve-filtering`
**Created**: 2026-01-31
**Status**: Draft
**Input**: User description: "Implement Sieve Filtering (RFC 5228) for server-side email rules like vacation responses and folder sorting."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Rules Engine & FileInto (Priority: P1)

As a user, I want my incoming emails to be automatically sorted into folders based on Subject or Sender, so that my Inbox remains organized without client-side rules running.

**Why this priority**: Core functionality of Sieve. Enables basic organization.

**Independent Test**: Can be tested by injecting a message with specific headers and verifying it appears in a subfolder instead of Inbox.

**Acceptance Scenarios**:

1. **Given** a rule "If Subject contains 'Bills', file into 'Finance'", **When** an email arrives with Subject "Your Monthly Bills", **Then** the message is delivered to the 'Finance' mailbox.
2. **Given** a rule "If Sender is 'bad@spam.com', discard", **When** such an email arrives, **Then** it is accepted (250 OK) but silently deleted (not delivered).
3. **Given** no active script, **When** email arrives, **Then** it delivers to Inbox.
4. **Given** a script with a syntax error, **When** email arrives, **Then** fall back to Implicit Keep (Inbox) and log the error.

---

### User Story 2 - Vacation Auto-Reply (Priority: P2)

As a user, I want the server to automatically reply to senders when I am out of office, so they know I'm unavailable.

**Why this priority**: High-value feature for business users.

**Independent Test**: Send an email to the user; verify an automatic reply is received. Send a second email immediately; verify NO reply (rate limiting).

**Acceptance Scenarios**:

1. **Given** an active Vacation rule, **When** a new sender emails me, **Then** they receive an automatic reply with my custom message.
2. **Given** I have already replied to 'sender@example.com' today, **When** they email me again, **Then** no second reply is sent (until the :days period expires).
3. **Given** a mailing list message (Precedence: list/bulk), **When** received, **Then** no vacation reply is sent (to prevent loops).

---

### User Story 3 - Script Management (Priority: P3)

As a developer/admin, I want to manage Sieve scripts via API, so I can integrate this into the Web Admin UI.

**Why this priority**: Necessary to enable the feature, but the engine (P1) is the hard part.

**Independent Test**: Upload a script via API, then verify it is active for the next email delivery.

**Acceptance Scenarios**:

1. **Given** a valid Sieve script text, **When** uploaded via POST /api/v1/users/{id}/sieve, **Then** it validates syntax and saves successfully.
2. **Given** an invalid script, **When** uploaded, **Then** the API returns a 400 bad request with parse errors.
3. **Given** multiple scripts, **When** calling PUT /active, **Then** the specified script becomes the active filter.

### User Story 4 - ManageSieve Integration (Priority: P3)

As a desktop mail user (Thunderbird), I want to update my filters directly from my email client using the ManageSieve protocol, so I don't need a separate web login.

**Why this priority**: Provides true "Drop-in" compatibility with standard clients.

**Independent Test**: Connect using `sieve-connect` or Thunderbird to port 4190, upload a script, and verify it is saved.

**Acceptance Scenarios**:

1. **Given** a ManageSieve client, **When** connecting to port 4190, **Then** the server challenges for SASL authentication.
2. **Given** an authenticated session, **When** sending `PUTSCRIPT "vacation" {content...}`, **Then** the script is stored.
3. **Given** an active script, **When** sending `SETACTIVE "vacation"`, **Then** the script becomes the active filter.

### Edge Cases

- **Infinite Loops**: System prevents forwarding loops (e.g., A -> B -> A) via header tracking.
- **Quota Exceeded**: If `fileinto` fails due to quota (if implemented), message MUST fall back to Inbox or Soft Fail (4xx), never lost.
- **Script Limits**: Scripts must be limited in size (e.g., 32KB) and execution steps to prevent DoS.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST implement a Sieve interpreter compliant with RFC 5228.
- **FR-002**: System MUST support the `fileinto` action (RFC 5228) to deliver to specific mailboxes.
- **FR-003**: System MUST support the `vacation` extension (RFC 5230) with duplicate tracking (replied-to database).
- **FR-004**: System MUST hook Sieve processing into the Local Delivery Agent (LDA) phase (after Spam/AV, before storage).
- **FR-005**: System MUST fail open (Implicit Keep) if script execution fails (runtime error).
- **FR-006**: System MUST provide REST API endpoints to Upload, List, Get, Delete, and Activate scripts per user.
- **FR-007**: System MUST implement the ManageSieve Protocol (RFC 5804) on TCP port 4190, supporting `PUTSCRIPT`, `SETACTIVE`, `DELETESCRIPT`, `GETSCRIPT`, `LISTSCRIPTS`, and SASL authentication.

### Key Entities *(include if feature involves data)*

- **SieveScript**: Key=(UserID, ScriptName), Attributes=(Content, IsActive, CreatedAt).
- **VacationTracker**: Key=(UserID, SenderAddress, RuleHash), Attributes=(LastRepliedAt).

### Assumptions

- **Existing Delivery**: We have a clear point in `smtp` or `queue` processing where local delivery happens.
- **Mailbox Creation**: `fileinto` will automatically create the folder if it doesn't exist (commonly requested behavior), or fallback to Inbox if strict (standard says implementation defined). We assume **automatic creation** for better UX.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: "FileInto" rules successfully route messages 100% of the time.
- **SC-002**: Vacation replies are never sent more than once per configured period (default 7 days) to the same sender.
- **SC-003**: Sieve processing adds less than 50ms latency to delivery on average.
- **SC-004**: Malformed scripts uploaded via API are rejected 100% of the time (validation at upload).
