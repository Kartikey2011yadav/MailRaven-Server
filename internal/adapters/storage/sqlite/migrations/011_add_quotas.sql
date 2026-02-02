-- Add storage quota columns to users table
ALTER TABLE users ADD COLUMN storage_quota INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN storage_used INTEGER DEFAULT 0;
