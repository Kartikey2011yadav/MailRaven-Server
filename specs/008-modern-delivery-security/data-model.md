# Data Model: Modern Delivery Security

## Entities

### TLS Report (`TLSReport`)

Represents an aggregate report received from a remote MTA (like Google or Microsoft) about their ability to send email to us securely.

| Field | Type | Description |
|-------|------|-------------|
| ID | UUID | Unique Identifier |
| ReportID | String | External ID from the sender |
| Provider | String | Organization sending the report (e.g., "google.com") |
| StartDate | Timestamp | Reporting period start |
| EndDate | Timestamp | Reporting period end |
| TotalCount | Integer | Total sessions attempted |
| SuccessCount | Integer | Successful secure sessions |
| FailureCount | Integer | Failed sessions |
| RawJSON | JSON/Text | Full report body for detailed debugging |
| CreatedAt | Timestamp | Ingestion time |

## Storage Schema (SQLite/Postgres)

### Table: `tls_reports`

```sql
CREATE TABLE tls_reports (
    id TEXT PRIMARY KEY,
    report_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    total_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    raw_json TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tls_reports_date ON tls_reports(start_date);
CREATE INDEX idx_tls_reports_provider ON tls_reports(provider);
```

## DANE Verification Cache (InMemory)

No database storage required for DANE. We will use an in-memory TTL cache for TLSA records to avoid DNS lookups on every single message if possible, or just rely on the OS/Recursive resolver's cache.
