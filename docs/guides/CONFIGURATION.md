# Configuration Reference

MailRaven is configured via a YAML file, typically located at `/etc/mailraven/config.yaml`. All values can be overridden via environment variables (see table below).

## Top Level

| Key | Type | Description |
|-----|------|-------------|
| `domain` | string | The primary domain for the mail server (e.g., `example.com`). |
| `mode` | string | Deployment mode: `standalone` (default), `docker`, or `kubernetes`. |
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

## Redis (Distributed Mode)

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `redis.enabled` | bool | `false` | Enable Redis for distributed caching, rate limiting, and pub/sub. |
| `redis.addr` | string | - | Redis server address (e.g., `redis:6379`). |
| `redis.password` | string | `""` | Redis authentication password. |
| `redis.db` | int | `0` | Redis database number. |

## NATS (Message Broker)

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `nats.enabled` | bool | `false` | Enable NATS JetStream for async task distribution. |
| `nats.url` | string | - | NATS server URL (e.g., `nats://nats:4222`). |

## Object Store

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `object_store.driver` | string | `disk` | Blob storage backend: `disk` or `minio`. |
| `object_store.endpoint` | string | - | MinIO server endpoint (e.g., `minio:9000`). |
| `object_store.bucket` | string | `mailraven-blobs` | MinIO bucket name. |
| `object_store.access_key` | string | - | MinIO access key. |
| `object_store.secret_key` | string | - | MinIO secret key. |
| `object_store.use_ssl` | bool | `false` | Use HTTPS for MinIO connection. |

## Environment Variable Overrides

All critical config values can be set via environment variables. Env vars take precedence over YAML.

| Environment Variable | Config Key | Description |
|---|---|---|
| `MAILRAVEN_DOMAIN` | `domain` | Primary mail domain |
| `MAILRAVEN_MODE` | `mode` | Deployment mode |
| `MAILRAVEN_SMTP_HOSTNAME` | `smtp.hostname` | SMTP EHLO hostname |
| `MAILRAVEN_JWT_SECRET` | `api.jwt_secret` | JWT signing secret |
| `MAILRAVEN_CORS_ORIGINS` | `api.cors_origins` | Comma-separated allowed origins |
| `MAILRAVEN_STORAGE_DSN` | `storage.dsn` | PostgreSQL connection string |
| `MAILRAVEN_STORAGE_DB_PATH` | `storage.db_path` | SQLite file path |
| `MAILRAVEN_STORAGE_BLOB_PATH` | `storage.blob_path` | Blob storage directory |
| `MAILRAVEN_DKIM_KEY_PATH` | `dkim.private_key_path` | DKIM private key path |
| `MAILRAVEN_REDIS_ADDR` | `redis.addr` | Redis address (also sets `redis.enabled=true`) |
| `MAILRAVEN_REDIS_PASSWORD` | `redis.password` | Redis password |
| `MAILRAVEN_NATS_URL` | `nats.url` | NATS URL (also sets `nats.enabled=true`) |
| `MAILRAVEN_OBJECT_STORE_DRIVER` | `object_store.driver` | `disk` or `minio` |
| `MAILRAVEN_OBJECT_STORE_ENDPOINT` | `object_store.endpoint` | MinIO endpoint |
| `MAILRAVEN_OBJECT_STORE_BUCKET` | `object_store.bucket` | MinIO bucket |
| `MAILRAVEN_OBJECT_STORE_ACCESS_KEY` | `object_store.access_key` | MinIO access key |
| `MAILRAVEN_OBJECT_STORE_SECRET_KEY` | `object_store.secret_key` | MinIO secret key |
