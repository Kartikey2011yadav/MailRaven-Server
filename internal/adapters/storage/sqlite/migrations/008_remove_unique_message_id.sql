-- Migration: 008_remove_unique_message_id.sql
-- Description: Remove UNIQUE constraint on message_id to allow multi-recipient limits and IMAP COPY

DROP INDEX IF EXISTS idx_messages_message_id;
CREATE INDEX idx_messages_message_id ON messages(message_id);
