-- Add ACL column to mailboxes table for RFC 4314 support
ALTER TABLE mailboxes ADD COLUMN acl TEXT DEFAULT '{}';
