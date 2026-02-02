# Quickstart: Protocol Features

## CLI Management

MailRaven exposes CLI commands for managing administrative features like Quotas.

### Managing Quotas

```bash
# Set quota for user@example.com to 1GB
mailraven users set-quota user@example.com --limit 1GB

# Check quota usage
mailraven users info user@example.com
```

### Managing ACLs (Shared Folders)

Currently, ACLs are primarily managed via IMAP clients that support the ACL extension (like Thunderbird with plugins, or scripts).

Using `openssl s_client` (IMAP):

```text
a001 LOGIN admin password
a002 SETACL User/Folder friend@example.com lr
a003 GETACL User/Folder
```

## Mobile Context

For developers building the mobile client, fetch the agent context:

```bash
curl http://localhost:8080/api/mobile/context
```

This endpoint returns the verified routes available for the mobile app, including:
-   Auth endpoints (JWT)
-   Mailbox sync endpoints (JMAP-like or REST)
-   Quota status
