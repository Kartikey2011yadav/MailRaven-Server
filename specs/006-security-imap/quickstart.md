# Quickstart for Feature 006: Security & IMAP Groundwork

This feature adds spam protection (DNSBL, Rspamd) and a basic IMAP4rev1 listener.

## Prerequisites

- **Rspamd**: A running instance of Rspamd is required for content filtering.
  - Docker: `docker run -p 11333:11333 rspamd/rspamd`
- **DNS**: Outbound DNS access (UDP/53) is required for DNSBL checks.

## Configuration

Add the `security` and `imap` sections to your `config.yaml`.

```yaml
server:
  port: 2525
  domain: "localhost"

# Feature 006: Security Configuration
security:
  spam:
    enabled: true
    dnsbl_zones:
      - "zen.spamhaus.org"
      - "bl.spamcop.net"
    rspamd_url: "http://localhost:11333"
    reject_threshold: 7.0   # Score above this rejects the email
    add_header_threshold: 4.0 # Score above this adds X-Spam-Flag

# Feature 006: IMAP Configuration
imap:
  enabled: true
  port: 1143           # Non-privileged port for testing
  tls:
    enabled: false      # Set true if certs provided
    cert_file: ""
    key_file: ""
```

## Running the Server

1.  **Start Rspamd (Optional but recommended)**:
    ```bash
    docker run -d -p 11333:11333 --name rspamd rspamd/rspamd
    ```

2.  **Run MailRaven**:
    ```bash
    go run cmd/server/main.go
    ```

## Verification Scenarios

### 1. IMAP Capabilities Check
Use `telnet` or `netcat` to connect to the IMAP port.

```bash
telnet localhost 1143
```

**Expected Interaction:**
```
* OK [CAPABILITY IMAP4rev1 AUTH=PLAIN STARTTLS] MailRaven IMAP Server Ready
A01 CAPABILITY
* CAPABILITY IMAP4rev1 AUTH=PLAIN STARTTLS
A01 OK CAPABILITY completed
A02 LOGOUT
* BYE Logging out
A02 OK LOGOUT completed
```

### 2. DNSBL Test (Simulated)
You cannot easily simulate a blocked IP from localhost without mocking, but the logs should show the lookup attempt.

Sending an email from `127.0.0.2` (common test IP for DNSBL positives) requires low-level network manipulation or unit tests.
Use the provided `internal/adapters/spam/dnsbl_test.go` to verify logic.

### 3. Rspamd Integration
Send an email with the GTUBE test string (Generic Test for Unsolicited Bulk Email).

**Body content:**
```
XJS*C4JDBQADN1.NSBN3*2IDNEN*GTUBE-STANDARD-ANTI-UBE-TEST-EMAIL*C.34X
```

**Expected Result:**
- Server should reject the message with a 550 error if `reject_threshold` is met.
- Server logs should show Rspamd scoring.
