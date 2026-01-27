# Data Model: Production Hardening

## 1. Schema Changes

### `configuration` (Application Config)
We will extend the YAML configuration structure, not the database schema.

```yaml
# New Config Sections
tls:
  enabled: true
  acme:
    enabled: true
    email: "admin@example.com"
    domains: ["mail.example.com"]
    cache_dir: "/data/certs"

spam:
  dnsbls:
    - "zen.spamhaus.org"
    - "bl.spamcop.net"
  max_recipients: 50
  rate_limit:
    window: 1t
    count: 100

backup:
  location: "/data/backups"
  retention_days: 7
```

## 2. Persistence Strategy

### ACME Certificates
- **Storage**: Filesystem (`DirCache`)
- **Location**: `/data/certs` (Mounted Volume)
- **Format**: PEM encoded files managed by `autocert`

### Spam Prevention
- **State**: In-memory (Rate limit counters)
- **Persistence**: None (Reset on restart is acceptable for rate limits)

### Backups
- **Format**: 
  1. `meta.db`: SQLite DB snapshot (via `VACUUM INTO`)
  2. `blobs/`: Gzipped email bodies (Incremental sync or copy)
- **Manifest**: `backup-manifest.json` containing timestamp and version.
