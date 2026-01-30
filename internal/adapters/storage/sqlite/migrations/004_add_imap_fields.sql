-- Migration 004: Add IMAP fields to messages and create mailboxes table

-- Add columns to messages table
ALTER TABLE messages ADD COLUMN uid INTEGER DEFAULT 0;
ALTER TABLE messages ADD COLUMN mailbox TEXT DEFAULT 'INBOX';
ALTER TABLE messages ADD COLUMN flags TEXT DEFAULT '';
ALTER TABLE messages ADD COLUMN mod_seq INTEGER DEFAULT 0;

-- Create indexes for IMAP performance
CREATE INDEX idx_messages_mailbox_uid ON messages(recipient, mailbox, uid);

-- Create mailboxes table
CREATE TABLE mailboxes (
    name TEXT NOT NULL,
    user_id TEXT NOT NULL,
    uid_validity INTEGER NOT NULL,
    uid_next INTEGER DEFAULT 1,
    message_count INTEGER DEFAULT 0,
    PRIMARY KEY (user_id, name),
    FOREIGN KEY (user_id) REFERENCES users(email) ON DELETE CASCADE
);
