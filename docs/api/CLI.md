# MailRaven CLI

The `mailraven-cli` tool provides command-line access to the MailRaven Admin API.

## Installation

```bash
make build
# Binary is at bin/mailraven-cli
```

## Configuration

You can provide server and token via flags or environment variables.

```bash
export MAILRAVEN_ADMIN_TOKEN="<your-jwt-token>"
# Optional
export MAILRAVEN_API="http://localhost:8443"
```

To get an admin token, you must currently use the `/auth/login` endpoint with your admin credentials, or check server logs if an initial token is printed (not yet implemented).

## Commands

### Users

Manage user accounts.

- **List**: `mailraven-cli users list`
- **Create**: `mailraven-cli users create <email> <password> [role]`
- **Delete**: `mailraven-cli users delete <email>`
- **Role**: `mailraven-cli users role <email> <role>`

### System

- **Stats**: `mailraven-cli system stats` (Coming Soon)

## Example

```bash
./bin/mailraven-cli users create boss@example.com strictPassword admin
./bin/mailraven-cli users list
```
