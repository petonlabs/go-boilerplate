ALTER TABLE users
  ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS email TEXT,
  ADD COLUMN IF NOT EXISTS password_hash TEXT,
  ADD COLUMN IF NOT EXISTS password_reset_token TEXT,
  ADD COLUMN IF NOT EXISTS password_reset_expires TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS oauth_provider TEXT,
  ADD COLUMN IF NOT EXISTS oauth_provider_id TEXT,
  ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS deletion_scheduled_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS users_deleted_at_idx ON users (deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS users_email_idx ON users (email);

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS clerk_id TEXT,
  ADD COLUMN IF NOT EXISTS external_id TEXT,
  ADD COLUMN IF NOT EXISTS first_name TEXT,
  ADD COLUMN IF NOT EXISTS last_name TEXT,
  ADD COLUMN IF NOT EXISTS image_url TEXT,
  ADD COLUMN IF NOT EXISTS raw_payload JSONB DEFAULT '{}'::jsonb;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'users_external_id_unique_idx') THEN
    EXECUTE 'CREATE UNIQUE INDEX users_external_id_unique_idx ON users (lower(external_id)) WHERE external_id IS NOT NULL';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'users_clerk_id_unique_idx') THEN
    EXECUTE 'CREATE UNIQUE INDEX users_clerk_id_unique_idx ON users (lower(clerk_id)) WHERE clerk_id IS NOT NULL';
  END IF;
END$$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='password_reset_token') THEN
    UPDATE users SET password_reset_token = NULL
    WHERE password_reset_token IS NOT NULL
      AND password_reset_token ~ '^[0-9a-fA-F]{32}$';
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'oauth_provider_check') THEN
    ALTER TABLE users ADD CONSTRAINT oauth_provider_check CHECK (
      (oauth_provider IS NULL AND oauth_provider_id IS NULL) OR
      (oauth_provider IS NOT NULL AND oauth_provider_id IS NOT NULL)
    );
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'oauth_provider_unique') THEN
    ALTER TABLE users ADD CONSTRAINT oauth_provider_unique UNIQUE (oauth_provider, oauth_provider_id);
  END IF;
END$$;
