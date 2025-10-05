-- Initial schema setup
CREATE TABLE IF NOT EXISTS schema_version (
  version INT PRIMARY KEY,
  applied_at TIMESTAMPTZ DEFAULT now(),
  dirty BOOLEAN DEFAULT FALSE
);

-- Create users table as a baseline
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Ensure older databases that were created before `applied_at` existed get the column
ALTER TABLE schema_version
  ADD COLUMN IF NOT EXISTS applied_at TIMESTAMPTZ DEFAULT now();

-- Backfill any existing rows without applied_at
UPDATE schema_version SET applied_at = now() WHERE applied_at IS NULL;
