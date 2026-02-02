# IMAP Extensions

## Quota (RFC 2087)

### CAPABILITY
`QUOTA`

### Commands

-   `GETQUOTA "RootName"`
-   `GETQUOTAROOT "MailboxName"`
-   `SETQUOTA "RootName" (Resource Limit ...)`

### Responses

-   `* QUOTA "RootName" (Resource Usage Limit ...)`
-   `* QUOTAROOT "Mailbox" "RootName" ...`

## ACL (RFC 4314)

### CAPABILITY
`ACL`

### Commands

-   `SETACL "Mailbox" "Identifier" "Rights"`
-   `DELETEACL "Mailbox" "Identifier"`
-   `GETACL "Mailbox"`
-   `LISTRIGHTS "Mailbox" "Identifier"`
-   `MYRIGHTS "Mailbox"`

### Rights (Standard)
`l r s w i p c d a`
