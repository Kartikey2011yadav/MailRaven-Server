# MailRaven Technical Concepts Guide

Everything you need to understand how an email server works, explained through MailRaven's implementation.

---

## Part 1: How Email Works (The Big Picture)

### The Journey of an Email

```
Sender (Alice)                                    Recipient (Bob)
    |                                                 |
    v                                                 |
[Alice's Mail Client]                                 |
    |                                                 |
    | SMTP (port 25/587)                              |
    v                                                 |
[Alice's Mail Server] ---MX lookup--> DNS             |
    |                                                 |
    | SMTP (port 25, with STARTTLS)                   |
    v                                                 |
[Bob's Mail Server]  ← THIS IS MAILRAVEN             |
    |                                                 |
    | Store in DB + blob                              |
    |                                                 |
    |---- IMAP (port 143/993) --------> [Bob's Client]
    |---- REST API (port 8080) -------> [Bob's Mobile App]
```

### Protocols Involved

| Protocol | Port | Purpose | MailRaven File |
|----------|------|---------|----------------|
| SMTP | 25 | Receive mail from other servers | `internal/adapters/smtp/server.go` |
| SMTP | 587 | Submit mail from local clients | Same server, TLS required |
| IMAP | 143/993 | Clients read mail | `internal/adapters/imap/` |
| HTTP | 8080 | REST API for mobile/web | `internal/adapters/http/` |
| ManageSieve | 4190 | Clients manage filter rules | `internal/adapters/managesieve/` |
| DNS | 53 | MX lookup, SPF/DKIM/DMARC | Used by validators |

---

## Part 2: Email Authentication (Anti-Spoofing)

Email has no built-in authentication — anyone can claim to be anyone. These protocols fix that:

### SPF (Sender Policy Framework) — RFC 7208

**What it does**: Answers "Is this IP address allowed to send mail for this domain?"

**How it works**:
1. Receiving server (MailRaven) gets email from IP `1.2.3.4` claiming to be `from@example.com`
2. MailRaven looks up DNS TXT record for `example.com`
3. Record says: `v=spf1 ip4:1.2.3.0/24 mx -all`
4. MailRaven checks: Is `1.2.3.4` in `1.2.3.0/24`? → Yes = **Pass**

**SPF record mechanisms**:
- `ip4:1.2.3.0/24` — Allow this IP range
- `mx` — Allow the domain's MX servers
- `include:_spf.google.com` — Check another domain's SPF too
- `-all` — Reject everything else (hard fail)
- `~all` — Soft fail everything else (mark as suspicious)

**Implementation**: `internal/adapters/smtp/validators/spf.go`

---

### DKIM (DomainKeys Identified Mail) — RFC 6376

**What it does**: Cryptographic proof that the email wasn't tampered with in transit.

**How it works**:
1. Sending server signs the email with its private key
2. Adds a `DKIM-Signature` header with the signature
3. Receiving server (MailRaven) fetches the public key from DNS
4. Verifies the signature matches the email content

**The signing process** (done when MailRaven sends mail):
```
Email body → SHA256 hash → "bh=" (body hash)
Selected headers + body hash → RSA-SHA256 sign → "b=" (signature)
```

**The verification process** (done when MailRaven receives mail):
```
Fetch public key from: [selector]._domainkey.[domain] TXT record
Recompute body hash → compare with "bh="
Verify RSA signature on headers using public key
```

**Implementation**: 
- Signing: `internal/adapters/smtp/dkim/signer.go`
- Verification: `internal/adapters/smtp/validators/dkim.go`

---

### DMARC (Domain-based Message Authentication) — RFC 7489

**What it does**: Tells receiving servers what to do when SPF/DKIM fail.

**How it works**:
1. MailRaven checks SPF → Pass/Fail
2. MailRaven checks DKIM → Pass/Fail
3. MailRaven looks up `_dmarc.example.com` TXT record
4. Record says: `v=DMARC1; p=reject; rua=mailto:reports@example.com`
5. If BOTH SPF and DKIM fail → apply policy (`reject`, `quarantine`, or `none`)

**Alignment**: DMARC requires that the domain in SPF/DKIM matches the From: header domain.

**Implementation**: `internal/adapters/smtp/validators/dmarc.go`

---

### DANE (DNS-based Authentication of Named Entities) — RFC 7672

**What it does**: Pins TLS certificates to DNS using DNSSEC. Prevents man-in-the-middle attacks on SMTP connections.

**How it works**:
1. Before connecting to a remote server, MailRaven looks up TLSA records
2. TLSA record: `_25._tcp.mail.example.com` contains certificate fingerprint
3. When TLS connection is established, MailRaven verifies the cert matches the TLSA record
4. Only trusted if DNSSEC validates (AD bit in DNS response)

**Modes in MailRaven**:
- `off` — Don't check DANE
- `advisory` — Check and log, but don't fail delivery
- `enforce` — Reject connection if TLSA doesn't match

**Implementation**: `internal/adapters/smtp/validators/dane.go`

---

## Part 3: Spam Filtering

### Bayesian Classifier (Paul Graham's Algorithm)

**Concept**: Learn from past spam/ham to predict future messages.

**Training phase**:
- User marks email as spam → tokenize → increment spam count for each word
- User marks email as ham → tokenize → increment ham count for each word

**Classification phase**:
1. Tokenize the incoming email (split into words, deduplicate)
2. For each word, calculate: `P(spam|word) = spamCount/totalSpam ÷ (spamCount/totalSpam + hamCount/totalHam)`
3. Take the 15 most "interesting" words (furthest from 0.5)
4. Combine probabilities: `P(spam) = (p1×p2×...×p15) / (p1×p2×...×p15 + (1-p1)×(1-p2)×...×(1-p15))`
5. If P(spam) > threshold → reject or mark as spam

**Implementation**: `internal/adapters/spam/bayesian/`

---

### Greylisting

**Concept**: Legitimate servers retry delivery; spammers usually don't.

**Algorithm**:
1. First time seeing (IP, sender, recipient) tuple → reject with "try again in 5 minutes"
2. If same tuple retries after 5 minutes → allow through (whitelist)
3. If same tuple retries too soon → reject again
4. Entries expire after 24 hours

**Why it works**: Spam botnets send millions of emails and rarely retry. Legitimate servers always retry (RFC 5321 requires it).

**Implementation**: `internal/adapters/spam/greylist/`

---

### DNSBL (DNS Block Lists)

**Concept**: Crowdsourced lists of known spam-sending IPs.

**How it works**:
1. Incoming connection from IP `1.2.3.4`
2. Query: `4.3.2.1.zen.spamhaus.org` (reversed IP)
3. If DNS returns a result → IP is blacklisted → reject connection

**Implementation**: Part of spam filter in `internal/adapters/spam/`

---

## Part 4: Mail Filtering (Sieve) — RFC 5228

### What is Sieve?

A domain-specific language for email filtering. Users write rules like:

```sieve
require ["fileinto", "vacation"];

# Move newsletters to a folder
if header :contains "List-Id" "newsletter" {
    fileinto "Newsletters";
    stop;
}

# Auto-reply when on vacation
vacation :days 7 :subject "Out of Office"
    "I'm on vacation until Monday. Will reply when I return.";

# Move spam to Junk
if header :contains "X-Spam-Status" "Yes" {
    fileinto "Junk";
    stop;
}
```

### Sieve Commands in MailRaven

| Command | What it does |
|---------|-------------|
| `keep` | Deliver to INBOX (default) |
| `fileinto "Folder"` | Deliver to specific folder |
| `discard` | Delete the message |
| `vacation` | Send auto-reply (rate-limited) |
| `stop` | Stop processing rules |
| `if/elsif/else` | Conditional logic |

### Sieve Tests

| Test | What it checks |
|------|---------------|
| `header :contains "Subject" "urgent"` | Subject contains "urgent" |
| `header :is "From" "boss@corp.com"` | From exactly matches |
| `header :matches "To" "*@lists.*"` | Wildcard matching |
| `anyof (test1, test2)` | OR logic |
| `allof (test1, test2)` | AND logic |
| `not test` | Negation |

**Implementation**: `internal/adapters/sieve/`

---

## Part 5: IMAP Protocol — RFC 3501

### What is IMAP?

IMAP lets email clients (Thunderbird, Apple Mail, Outlook) access mail stored on the server. Unlike POP3 (which downloads and deletes), IMAP keeps mail on the server.

### IMAP Session Lifecycle

```
Client connects → Server sends greeting
Client: LOGIN user pass → Server: OK
Client: SELECT INBOX → Server: * 50 EXISTS, * 2 RECENT
Client: FETCH 1:* (FLAGS ENVELOPE) → Server: message list
Client: FETCH 5 BODY[] → Server: full message
Client: STORE 5 +FLAGS (\Seen) → Server: OK (mark as read)
Client: IDLE → Server: (waits, notifies of new mail)
Client: DONE → Server: OK (exits IDLE)
Client: LOGOUT → Server: BYE
```

### Key IMAP Concepts

| Concept | Meaning |
|---------|---------|
| **UID** | Unique ID per message within a mailbox (never reused) |
| **Flags** | `\Seen`, `\Answered`, `\Flagged`, `\Deleted`, `\Draft` |
| **IDLE** | Long-polling — server pushes new mail notifications |
| **ACL** | Per-mailbox access control (who can read/write/admin) |
| **QUOTA** | Storage limits per user |
| **STARTTLS** | Upgrade to encrypted connection |

**Implementation**: `internal/adapters/imap/`

---

## Part 6: MIME (Email Message Format) — RFC 2045-2049

### Email Structure

An email is NOT just text. It's a structured document:

```
From: alice@example.com
To: bob@example.com
Subject: Hello
Content-Type: multipart/mixed; boundary="----boundary123"
DKIM-Signature: v=1; a=rsa-sha256; ...

------boundary123
Content-Type: text/plain

Hello Bob, here's the document.

------boundary123
Content-Type: application/pdf; name="report.pdf"
Content-Disposition: attachment; filename="report.pdf"
Content-Transfer-Encoding: base64

JVBERi0xLjQKMSAwIG9iago8PAovVHlwZSA...

------boundary123--
```

### MIME Concepts

| Part | Purpose |
|------|---------|
| Headers | Metadata (From, To, Subject, Content-Type) |
| Boundary | Separator between parts in multipart messages |
| text/plain | Plain text body |
| text/html | HTML body |
| multipart/mixed | Contains text + attachments |
| multipart/alternative | Same content in multiple formats (text + html) |
| Content-Transfer-Encoding | How binary data is encoded (base64, quoted-printable) |

**Implementation**: `internal/adapters/smtp/mime/parser.go`

---

## Part 7: Outbound Delivery

### MX (Mail Exchange) Lookup

When MailRaven needs to send mail to `bob@gmail.com`:
1. DNS query: `dig MX gmail.com` → `gmail-smtp-in.l.google.com (priority 5)`
2. Connect to `gmail-smtp-in.l.google.com:25`
3. STARTTLS (encrypt connection)
4. MAIL FROM, RCPT TO, DATA, QUIT

### Retry Strategy (Exponential Backoff)

If delivery fails (server down, temporary error):
```
Attempt 1: retry in 1 minute
Attempt 2: retry in 5 minutes
Attempt 3: retry in 15 minutes
Attempt 4: retry in 1 hour
Attempt 5: retry in 6 hours
Attempt 6-10: retry in 12-24 hours
After 10 attempts: permanent failure (bounce)
```

**Implementation**: `internal/adapters/smtp/delivery.go`

---

## Part 8: Transport Security

### MTA-STS (Mail Transfer Agent Strict Transport Security) — RFC 8461

**What it does**: Tells sending servers "always use TLS when connecting to us."

Published via HTTPS at `https://mta-sts.example.com/.well-known/mta-sts.txt`:
```
version: STSv1
mode: enforce
mx: mail.example.com
max_age: 86400
```

Plus a DNS TXT record: `_mta-sts.example.com` → `v=STSv1; id=20240101;`

### TLS-RPT (TLS Reporting) — RFC 8460

**What it does**: Other servers send you reports about TLS failures when connecting to your server.

DNS record: `_smtp._tls.example.com` → `v=TLSRPTv1; rua=mailto:tls-reports@example.com`

MailRaven receives and stores these reports at `POST /.well-known/tlsrpt`.

**Implementation**: `internal/adapters/http/handlers/` (MTA-STS serving, TLS-RPT ingestion)

---

## Part 9: Distributed Systems Concepts (MailRaven at Scale)

### Why Distribute?

A single server can handle ~100 concurrent connections. For 10,000+, you need multiple instances. But multiple instances need coordination:

| Problem | Solution | MailRaven Implementation |
|---------|----------|-------------------------|
| Rate limiting across instances | Shared counter | Redis sliding window |
| IMAP IDLE notifications | Cross-instance pub/sub | Redis Pub/Sub |
| Queue processing without duplicates | Distributed lock | PostgreSQL `FOR UPDATE SKIP LOCKED` |
| Shared blob storage | Object storage | MinIO (S3-compatible) |
| Async work distribution | Message queue | NATS JetStream |

### Scale-to-Zero (KEDA)

When no mail is arriving and no users are connected:
1. KEDA watches metrics (`mailraven_active_smtp_connections == 0`)
2. After 5 minutes idle → scales pods to 0
3. New SMTP connection arrives → KEDA triggers scale to 1
4. Pod starts in ~10 seconds → handles the connection

---

## Part 10: Security Checklist

| Layer | Protection | Implementation |
|-------|-----------|----------------|
| Transport | STARTTLS on SMTP/IMAP | TLS 1.2+ minimum |
| Authentication | JWT tokens (7-day expiry) | HS256 signing |
| Passwords | bcrypt hashing | `golang.org/x/crypto/bcrypt` |
| Input | Request body limits (10MB) | `http.MaxBytesReader` |
| Rate limiting | 100 req/min per IP | Sliding window counter |
| Email auth | SPF + DKIM + DMARC | Reject on DMARC fail |
| DNS | DANE for outbound TLS | TLSA record verification |
| Injection | Parameterized SQL queries | `$1, $2` placeholders |
| XSS | DOMPurify on email HTML | Client-side sanitization |
| CORS | Configurable origins | Middleware with allowlist |
