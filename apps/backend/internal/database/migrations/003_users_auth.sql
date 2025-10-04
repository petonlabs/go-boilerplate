-- Add production auth columns to users
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS password_hash TEXT,
  ADD COLUMN IF NOT EXISTS password_reset_token TEXT,
  ADD COLUMN IF NOT EXISTS password_reset_expires TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS oauth_provider TEXT,
  ADD COLUMN IF NOT EXISTS oauth_provider_id TEXT,
  ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS deletion_scheduled_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS users_deleted_at_idx ON users (deleted_at);

---- create above / drop below ----

-- Down: remove columns
DROP INDEX IF EXISTS users_deleted_at_idx;

ALTER TABLE users
  DROP COLUMN IF EXISTS email_verified,
  DROP COLUMN IF EXISTS password_hash,
  DROP COLUMN IF EXISTS password_reset_token,
  DROP COLUMN IF EXISTS password_reset_expires,
  DROP COLUMN IF EXISTS oauth_provider,
  DROP COLUMN IF EXISTS oauth_provider_id,
  DROP COLUMN IF EXISTS last_login_at,
  DROP COLUMN IF EXISTS deleted_at,
  DROP COLUMN IF EXISTS deletion_scheduled_at;
