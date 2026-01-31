# Quickstart: Testing Advanced Spam Filtering

**Feature**: 009-advanced-spam-filtering

How to manually verify the Spam Filter and Greylisting.

## 1. Configuration

Ensure `config/config.yaml` has the spam module enabled (if applicable).
By default, the feature comes enabled with reasonable defaults.

## 2. Testing Greylisting (Manual)

**Goal**: Verify unknown senders are temporarily rejected.

1.  **Start Server**: `go run cmd/mailraven/serve.go`
2.  **Connect**: `telnet localhost 2525`
3.  **Simulate New Sender**:
    ```text
    EHLO localhost
    MAIL FROM:<stranger@example.com>
    RCPT TO:<user@test.local>
    ```
4.  **Expected Response**:
    ```text
    451 Greylisted, please try again in 5 minutes
    ```
5.  **Wait**: Wait 5 minutes.
6.  **Retry**: Repeat step 3.
7.  **Expected Response**: `250 OK`

## 3. Testing Bayesian Filter

**Goal**: Verify content is scored.

1.  **Send Ham**:
    Send an email with "Meeting tomorrow lunch" to `user@test.local`.
    Check headers: `X-Spam-Status: No`
2.  **Send Spam**:
    Send an email with "Viagra casino money" to `user@test.local`.
    *Note: Initial system is untrained, so it might pass.*
3.  **Train (Simulated)**:
    Only accessible via code currently or IMAP move.
    *Development Helper*: Use `go test ./internal/adapters/spam/bayes` to feed a training corpus.
4.  **Verify**:
    Check headers: `X-Spam-Status: Yes`
    Check logs: `MsgID=... SpamScore=0.99 Action=Junk`

## 4. Database Inspection

Token counts can be inspected via sqlite3 CLI:
```bash
sqlite3 mailraven.db "SELECT * FROM bayes_tokens ORDER BY spam_count DESC LIMIT 5;"
sqlite3 mailraven.db "SELECT * FROM greylist;"
```
