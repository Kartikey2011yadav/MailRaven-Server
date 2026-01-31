# MailRaven Production Guide

This guide details how to deploy MailRaven in a production environment.

## Prerequisites

- **Domain Name**: A valid domain (e.g., `mail.example.com`).
- **Access**: Server with specialized ports open (25, 80, 443).
- **DNS**: Access to set MX, TXT, and A/AAAA records.

## Installation Options

### Option A: Storage Backend (PostgreSQL vs SQLite)

MailRaven supports two storage backends:

1. **SQLite (Default)**: Best for single-node, low-to-medium volume, simple backups.
2. **PostgreSQL**: Best for high volume, horizontal scaling (future), and robust data integrity.

To use PostgreSQL, set `storage.driver: postgres` in `config.yaml`.

### Option B: Deployment Method

#### 1. Binary Deployment (Linux)

1. **Build**:
   ```bash
   ./scripts/setup.sh
   # Result: bin/mailraven
   ```
2. **Configure**:
   Copy `config.example.yaml` to `/etc/mailraven/config.yaml` and edit.
3. **Run**:
   ```bash
   sudo ./bin/mailraven serve --config /etc/mailraven/config.yaml
   ```

#### 2. Docker Compose

1. **Edit Config**:
   Update `docker-compose.yml` environment variables or mount a config file.
2. **Run**:
   ```bash
   docker-compose up -d
   ```

## Configuration

See `config.example.yaml` for full options.

### Key Production Settings

- **TLS**: Enable `tls.acme` for automatic Let's Encrypt certificates.
- **DKIM**: Generate keys using `mailraven-cli gen-dkim` and publish the DNS record.
- **Spam**: Configure `spam.dnsbls` to reject known spammers (e.g., `zen.spamhaus.org`).

## DNS Configuration (Crucial)

| Type | Name | Value | Purpose |
|------|------|-------|---------|
| A    | mail | <IP_ADDRESS> | Server IP |
| MX   | @    | 10 mail.example.com. | Mail routing |
| TXT  | @    | "v=spf1 mx -all" | SPF (Allow only this server) |
| TXT  | default._domainkey | <PUBLIC_KEY_CONTENT> | DKIM Signature |
| TXT  | _dmarc | "v=DMARC1; p=quarantine; rua=mailto:postmaster@example.com" | DMARC Policy |
| A    | mta-sts | <IP_ADDRESS> | MTA-STS (Hosting Policy) |
| TXT  | _mta-sts | "v=STSv1; id=2024010101;" | MTA-STS (Version ID) |
| TXT  | _smtp._tls | "v=TLSRPTv1; rua=mailto:tls-reports@example.com" | TLS Reporting |

## Monitoring

- **Logs**: JSON formatted logs available (set `logging.format: json`).
- **Metrics**: HTTP endpoint available at `/api/v1/admin/stats`.

## Backup

- **SQLite**: Automatic `VACUUM INTO` backups if configured.
- **PostgreSQL**: Use standard `pg_dump` or the built-in backup service.
- **Blobs**: Backup the `data/blobs` directory regularly.

