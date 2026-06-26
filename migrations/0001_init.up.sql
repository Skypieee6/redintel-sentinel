-- 0001_init.up.sql
-- Baseline schema for RedIntel Sentinel.
-- This migration establishes core extensions and a schema_metadata table used
-- to record platform-level information. Feature tables (projects, assets, etc.)
-- are introduced in subsequent migrations.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS schema_metadata (
    key         TEXT PRIMARY KEY,
    value       TEXT NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO schema_metadata (key, value)
VALUES ('platform', 'redintel-sentinel')
ON CONFLICT (key) DO NOTHING;
