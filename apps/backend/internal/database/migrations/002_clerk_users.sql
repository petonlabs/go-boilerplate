-- Add Clerk/identity fields to users for webhook sync
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS clerk_id TEXT,
  ADD COLUMN IF NOT EXISTS external_id TEXT,
  ADD COLUMN IF NOT EXISTS first_name TEXT,
  ADD COLUMN IF NOT EXISTS last_name TEXT,
  ADD COLUMN IF NOT EXISTS image_url TEXT,
  ADD COLUMN IF NOT EXISTS raw_payload JSONB DEFAULT '{}'::jsonb;

CREATE UNIQUE INDEX IF NOT EXISTS users_external_id_idx ON users (external_id);
CREATE UNIQUE INDEX IF NOT EXISTS users_clerk_id_idx ON users (clerk_id);

---- create above / drop below ----

-- Down: remove columns and indexes
DROP INDEX IF EXISTS users_external_id_idx;
DROP INDEX IF EXISTS users_clerk_id_idx;

ALTER TABLE users
  DROP COLUMN IF EXISTS clerk_id,
  DROP COLUMN IF EXISTS external_id,
  DROP COLUMN IF EXISTS first_name,
  DROP COLUMN IF EXISTS last_name,
  DROP COLUMN IF EXISTS image_url,
  DROP COLUMN IF EXISTS raw_payload;
