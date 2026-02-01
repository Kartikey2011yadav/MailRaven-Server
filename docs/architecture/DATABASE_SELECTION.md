# Database Selection Guide

MailRaven supports both **SQLite** and **PostgreSQL** as storage backends. This guide explains how to configure your desired database mode.

## Quick Summary

| Database | Best For | Configuration Key |
|----------|----------|-------------------|
| **PostgreSQL** | Production, High Volume, Docker | `driver: postgres` |
| **SQLite** | Testing, Low Volume, Single Node | `driver: sqlite` |

## Docker Setup (Default)

The default `docker-compose.yml` is pre-configured for **PostgreSQL**.

### Using PostgreSQL (Recommended)
No changes needed. The `backend` service is configured with:

```yaml
environment:
  - MAILRAVEN_STORAGE_DRIVER=postgres
  - MAILRAVEN_STORAGE_DSN=postgres://mailraven:secretpassword@db:5432/mailraven?sslmode=disable
depends_on:
  db:
    condition: service_healthy
```

The `db` service runs a PostgreSQL 15 instance.

### Switching to SQLite in Docker
To use SQLite instead of Postgres in Docker:

1.  Open `docker-compose.yml`.
2.  Update the `backend` environment variables:
    ```yaml
    environment:
      - MAILRAVEN_STORAGE_DRIVER=sqlite
      - MAILRAVEN_STORAGE_DB_PATH=/data/mailraven.db
      # Remove MAILRAVEN_STORAGE_DSN
    ```
3.  Remove the `depends_on` section for `db`.
4.  (Optional) Remove the `db` service definition.

## Manual Configuration (`config.yaml`)

If running the binary directly, configure the `storage` section in your `config.yaml`.

### PostgreSQL
```yaml
storage:
  driver: "postgres"
  dsn: "postgres://user:pass@localhost:5432/mailraven?sslmode=disable"
  blob_path: "./data/blobs"
```

### SQLite
```yaml
storage:
  driver: "sqlite"
  db_path: "./data/mailraven.db"
  blob_path: "./data/blobs"
```
