DROP INDEX IF EXISTS idx_messages_recipient_starred;
DROP INDEX IF EXISTS idx_messages_recipient_mailbox;

ALTER TABLE messages DROP COLUMN IF EXISTS is_starred;
ALTER TABLE messages DROP COLUMN IF EXISTS modseq;
ALTER TABLE messages DROP COLUMN IF EXISTS flags;
ALTER TABLE messages DROP COLUMN IF EXISTS uid;
ALTER TABLE messages DROP COLUMN IF EXISTS mailbox;
