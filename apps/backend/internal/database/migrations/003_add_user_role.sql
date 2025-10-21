-- 003_add_user_role.sql

ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT;
