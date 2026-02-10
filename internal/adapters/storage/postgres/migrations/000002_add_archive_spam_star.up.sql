ALTER TABLE messages ADD COLUMN IF NOT EXISTS mailbox TEXT DEFAULT 'INBOX';
ALTER TABLE messages ADD COLUMN IF NOT EXISTS uid BIGINT DEFAULT 0;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS flags TEXT DEFAULT '';
ALTER TABLE messages ADD COLUMN IF NOT EXISTS modseq BIGINT DEFAULT 0;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS is_starred BOOLEAN DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_messages_recipient_mailbox ON messages(recipient, mailbox);
CREATE INDEX IF NOT EXISTS idx_messages_recipient_starred ON messages(recipient, is_starred);
