# Quickstart: Testing Feature 008

## Prerequisites

- Local DNS spoofing (hosts file) to simulate `mta-sts.localhost`.
- `curl` tool.
- Running MailRaven dev server.

## 1. Test MTA-STS Policy

1.  Add to local hosts file (or use `curl --resolve`):
    `127.0.0.1 mta-sts.localhost`

2.  Run the request:
    ```bash
    curl -k -H "Host: mta-sts.localhost" https://127.0.0.1:8443/.well-known/mta-sts.txt
    ```

    **Expected Output**:
    ```text
    version: STSv1
    mode: testing
    mx: localhost
    max_age: 86400
    ```

## 2. Test TLS-RPT Ingestion

Send a sample report:

```bash
curl -k -X POST https://127.0.0.1:8443/.well-known/tlsrpt \
    -H "Content-Type: application/tlsrpt+json" \
    -d '{
      "organization-name": "Google",
      "date-range": {
        "start-datetime": "2026-01-01T00:00:00Z",
        "end-datetime": "2026-01-01T23:59:59Z"
      },
      "contact-info": "smtp-tls-reporting@google.com",
      "report-id": "2026-01-01-google-mailraven",
      "policies": []
    }'
```

**Verify DB**:
```bash
sqlite3 data/mailraven.db "SELECT * FROM tls_reports;"
```

## 3. Test DANE Outbound (Integration Test only)

Since DANE requires live DNSSEC, manual testing is hard locally. Run the provided integration test:

```bash
go test -v ./internal/adapters/smtp/validators/ -run TestDANE
```
