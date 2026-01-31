# Quickstart: Sieve Filtering

## Overview
This feature adds Sieve (RFC 5228) support, allowing users to filter incoming emails into folders, reject them, or send vacation auto-replies.

## Prerequisites
- **Client**: Any ManageSieve-compatible client (e.g., Thunderbird, Roundcube) or use the Web Admin API.
- **Port**: Ensure TCP 4190 is exposed.

## Testing ManageSieve
You can use `openssl` or a dedicated client.

```bash
# Connect
openssl s_client -connect localhost:4190 -starttls sieve

# Authenticate
# (Base64 encoded "\0username\0password")
AUTHENTICATE "PLAIN" "AHNvbWUudXNlcgBzZWNyZXQ="

# Upload Script
PUTSCRIPT "myscript" {61+}
require "fileinto";
if header :contains "subject" "spam" {
  fileinto "Junk";
}

# Activate
SETACTIVE "myscript"
```

## Running Tests
Run the unit and integration tests to verify the Sieve engine processes emails correctly.

```bash
go test ./internal/adapters/sieve/... ./internal/adapters/managesieve/...
```
