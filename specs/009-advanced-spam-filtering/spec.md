# Feature Specification: Advanced Spam Filtering

**Feature Branch**: `009-advanced-spam-filtering`
**Created**: 2026-01-31
**Status**: Draft
**Input**: User description: "Implement Advanced Spam Filtering (Bayesian & Greylisting) to close the strategic gap with Mox."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Greylisting Protection (Priority: P1)

As an administrator, I want the server to temporarily reject emails from unknown senders so that simpler spam bots (which don't retry) are blocked before they deliver payload.

**Why this priority**: High impact on volume reduction with low implementation complexity. Stops "fire and forget" spam.

**Independent Test**: Can be tested by sending an email from a new IP; immediate delivery should fail with 4xx, retry after delay should succeed.

**Acceptance Scenarios**:

1. **Given** an unknown sender/IP/recipient tuple, **When** they attempt delivery, **Then** the server responds with a 4xx temporary error (e.g., 451 Graylisted).
2. **Given** a sender who has previously successfully retried (whitelisted), **When** they attempt delivery, **Then** the server accepts the message immediately.
3. **Given** a retrying sender after the "retry interval", **When** they attempt delivery again, **Then** the server accepts the message and records the tuple as valid.

---

### User Story 2 - Bayesian Content Analysis (Priority: P2)

As a user, I want incoming emails to be scanned and classified based on content patterns so that spam ends up in my Junk folder automatically.

**Why this priority**: Provides the primary "smart" filtering layer for content that passes structural checks (SPF/DKIM).

**Independent Test**: Can be tested by feeding known spam/ham samples and verifying the resulting classification score headers or folder placement.

**Acceptance Scenarios**:

1. **Given** an incoming email with "spammy" keywords (matching trained tokens), **When** processed, **Then** it receives a high spam probability score.
2. **Given** an email with a high spam score (above threshold), **When** delivered to the mailbox, **Then** it is automatically placed in the Junk folder.
3. **Given** an email with a low spam score, **When** delivered, **Then** it is placed in the Inbox.

---

### User Story 3 - Training via Feedback Loop (Priority: P3)

As a user, I want to teach the filter by moving emails between Inbox and Junk, so that it becomes more accurate over time.

**Why this priority**: Essential for accuracy improvement and correcting false positives/negatives.

**Independent Test**: Move a "ham" message to Junk, verify tokens are updated in the database.

**Acceptance Scenarios**:

1. **Given** a message in Inbox, **When** moved to Junk folder, **Then** the system learns it as "Spam" (increments spam counts for tokens).
2. **Given** a message in Junk, **When** moved to Inbox, **Then** the system learns it as "Ham" (increments ham counts for tokens).

### Edge Cases

- **Cold Start**: When the Bayesian database is empty, all mails are neutral (Ham).
- **Retries Too Fast**: If a sender retries before the Greylist delay expires, they receive another 4xx error.
- **Large Emails**: Messages over a certain size (e.g., 200KB) should skip Bayesian/Tokenization to preserve performance.

### Assumptions

- **Global Learning**: Training data is shared globally across the instance (consistent with Mox/Standard Self-Hosted).
- **Storage Availability**: The existing persistence layer can handle the additional load of token storage.
- **Client Support**: Clients support moving messages to a "Junk" folder to trigger feedback (standard IMAP behavior).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST implement a Greylisting mechanism in the SMTP reception pipeline that tracks (Sender IP, Sender Email, Recipient Email) tuples.
- **FR-002**: System MUST allow configuration of "Retry Delay", "Record Expiration", and "Whitelist/Exemptions" (e.g., SPF-pass bypass).
- **FR-003**: System MUST provide a Bayesian classifier engine capable of preventing "poisoning" and tokenizing email bodies/headers.
- **FR-004**: System MUST store Bayesian tokens and their Spam/Ham counts in the database.
- **FR-005**: System MUST score all incoming messages during the SMTP DATA phase and append `X-Spam-Status` and `X-Spam-Score` headers.
- **FR-006**: System MUST route messages with `X-Spam-Score` > Configured Threshold to the `Junk` folder during delivery.
- **FR-007**: System MUST detect IMAP message movements (Move/Copy) between "Inbox" (or others) and "Junk" to trigger retraining.
- **FR-008**: System MUST support a global training dictionary for all users on the instance (consistent with single-tenant self-hosted use case).

### Key Entities *(include if feature involves data)*

- **GreylistEntry**: Key=(IP, Sender, Recipient), Attributes=(FirstSeenAt, LastAllowedAt).
- **BayesToken**: Key=(Word/Token), Attributes=(SpamCount, HamCount, TotalCount).
- **SpamConfig**: Settings=(GreylistEnabled, BayesEnabled, SpamThreshold).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Greylisting rejects 100% of first-attempt connections from non-whitelisted tuples.
- **SC-002**: Valid retries (post-delay) are accepted 100% of the time.
- **SC-003**: Spam classification adds less than 100ms overhead to message processing time on average.
- **SC-004**: Moving a message to/from Junk triggers a database update for token counts (verifiable via log or db check).
