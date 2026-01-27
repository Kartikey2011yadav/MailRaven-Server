-- Add role column to users table
ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'user';

-- Set initial admin (optional, can be done via CLI later, or assume first user)
-- UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
