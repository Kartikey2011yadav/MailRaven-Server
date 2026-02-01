# Configuration Reference

MailRaven is configured via a YAML file, typically located at `/etc/mailraven/config.yaml`.

## Top Level

| Key | Type | Description |
|-----|------|-------------|
| `domain` | string | The primary domain for the mail server (e.g., `example.com`). |
| `smtp` | object | SMTP server settings. |
| `api` | object | API server settings. |
## Storage

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `driver` | string | `sqlite` | Database driver. Options: `sqlite`, `postgres`. |
| `db_path` | string | - | Path to SQLite DB file (required if driver is `sqlite`). |
| `dsn` | string | - | Postgres connection string (required if driver is `postgres`). |
| `blob_path` | string | - | Directory to store email bodies (blobs). |

## DKIM

| Key | Type | Default | Description |
| `spam` | object | Spam protection settings. |
| `backup` | object | Backup settings. |
| `logging` | object | Logging configuration. |

## SMTP

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `port` | int | `25` | Port to listen for incoming SMTP traffic. |
| `hostname` | string | (Required) | The hostname used in SMTP HELO/EHLO headers. |
| `max_size` | int | `10485760` | Maximum message size in bytes (default 10MB). |
| `dane.mode` | string | `advisory` | DANE verification mode for outbound mail. Options: `off`, `advisory` (log only), `enforce` (fail delivery on mismatch). |

## API

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `host` | string | `0.0.0.0` | IP to bind the HTTP server to. |
| `port` | int | `8443` | Port for the HTTP API and Web Admin. |
| `tls` | bool | `false` | Enable HTTPS (requires `tls_cert` and `tls_key` OR `tls.acme`). |
| `tls_cert` | string | - | Path to TLS certificate (PEM). |
| `tls_key` | string | - | Path to TLS private key (PEM). |
| `jwt_secret` | string | (Required) | Secret key for signing session tokens. |

## TLS & ACME (Automatic HTTPS)

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `false` | Global toggle for TLS features. |
| `acme.enabled` | bool | `false` | Enable automatic certificates via Let's Encrypt. |
| `acme.email` | string | - | Email address for Let's Encrypt registration. |
| `acme.domains` | list | - | List of domains to request certificates for. |
| `acme.cache_dir` | string | `/data/certs` | Directory to store ACME certificates. |

## Spam Protection

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `true` | Enable or disable spam protection. |
| `rspamd_url` | string | `http://localhost:11333` | URL of the Rspamd instance. |
| `dnsbls` | list | `[]` | List of DNSBL providers (e.g., `zen.spamhaus.org`). |
| `reject_score` | float | `15.0` | Score threshold to reject message (5xx). |
| `header_score` | float | `6.0` | Score threshold to add X-Spam headers. |
| `max_recipients` | int | `50` | Maximum number of recipients per message. |
| `rate_limit.count` | int | `100` | Max SMTP connections/commands per window. |
| `rate_limit.window` | string | `1h` | Duration of the rate limit window. |

## ManageSieve

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `port` | int | `4190` | Port to listen for ManageSieve connections. |
| `max_script_size` | int | `32768` | Maximum size of a Sieve script in bytes (default 32KB). |
| `vacation_min_days` | int | `1` | Minimum days between vacation replies to the same sender. |

## IMAP

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `true` | Enable or disable IMAP server. |
| `port` | int | `143` | Listener port for IMAP (STARTTLS supported). |
| `port_tls` | int | `993` | Listener port for IMAP over TLS (Implicit). |
| `allow_insecure_auth` | bool | `false` | Allow LOGIN command on unencrypted connection. |
| `tls_cert` | string | - | Path to TLS certificate. |
| `tls_key` | string | - | Path to TLS private key. |

## Backup

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `false` | Enable automated backups. |
| `schedule` | string | `0 2 * * *` | Cron expression for backup schedule. |
| `location` | string | - | Directory to store backups. |
| `retention_days` | int | `7` | Number of days to keep backups. |

## Storage

| Key | Type | Description |
|-----|------|-------------|
| `db_path` | string | Path to the SQLite database file. |
| `blob_path` | string | Path to the directory for storing email bodies/attachments. |

## DKIM

| Key | Type | Description |
|-----|------|-------------|
| `selector` | string | DKIM selector (e.g., `default`). |
| `private_key_path` | string | Path to the RSA private key for DKIM signing. |
