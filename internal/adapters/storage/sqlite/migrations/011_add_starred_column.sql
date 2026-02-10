-- Add is_starred column for marking important messages
-- Corresponds to IMAP \Flagged system flag

ALTER TABLE messages ADD COLUMN is_starred INTEGER NOT NULL DEFAULT 0;

-- Index for filtering starred messages efficiently
CREATE INDEX idx_messages_user_starred ON messages(recipient, is_starred);
