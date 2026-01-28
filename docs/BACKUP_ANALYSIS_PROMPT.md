# Backup Strategy Analysis Prompt

*Copy and use this prompt with an AI assistant to analyze your backup configuration and disaster recovery plan.*

---

**Context**: I am running MailRaven, a Go-based email server with a choice of SQLite or PostgreSQL storage.

**My Configuration**:
- Storage Driver: [SQLITE/POSTGRES]
- Blob Path: `[PATH_TO_BLOBS]`
- Backup Path: `[PATH_TO_BACKUPS]`

**Setup**:
1. **Database**: [Describe your DB size and transaction volume]
2. **Storage**: [Describe your disk type, e.g., local SSD, S3-mounted, etc.]

**Request**:
Please check my backup strategy for the following risks:
1. **Atomicity**: Are database and blob storage backups consistent? (MailRaven stores message metadata in DB and body in FS).
2. **Retention**: Am I keeping backups long enough for ransomware recovery?
3. **Offsite**: Is my strategy compliant with 3-2-1 backup rule?

**Current Implementation**:
- SQLite: Uses `VACUUM INTO` command.
- Postgres: Uses `pg_dump -Fc`.
- Blobs: File system copy.

**Question**:
What specific script or tool should I wrap around these primitives to ensure point-in-time recovery (PITR) consistency between the Database and the Blob store, considering that an email might arrive during the backup window?

---
