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
