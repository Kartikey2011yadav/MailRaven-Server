# Quickstart: Protocol Completion

## Prerequisites

-   Go 1.24+
-   MailRaven Server codebase

## Running Tests

### Quota Tests

```bash
go test -v ./mox/imapserver -run TestQuota
```

### ACL Tests

```bash
go test -v ./mox/imapserver -run TestACL
```

## Manual Verification

1.  **Start Server**: `go run ./mox/main.go serve`
2.  **Connect**: `openssl s_client -connect localhost:993`
3.  **Login**: `a001 LOGIN user pass`
4.  **Set Quota**: `a002 SETQUOTA "" (STORAGE 100000)`
5.  **Check ACL**: `a003 MYRIGHTS INBOX`
