-- 0006_assets.up.sql
CREATE TABLE IF NOT EXISTS assets (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type       TEXT NOT NULL CHECK (type IN
                 ('domain','subdomain','ip','cidr','asn','dns_record','certificate','technology')),
    value      TEXT NOT NULL,
    tags       TEXT[] NOT NULL DEFAULT '{}',
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','archived')),
    first_seen TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, type, value)
);
CREATE INDEX IF NOT EXISTS idx_assets_org ON assets(org_id);
CREATE INDEX IF NOT EXISTS idx_assets_project ON assets(project_id);
CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(type);
CREATE INDEX IF NOT EXISTS idx_assets_value ON assets(value);
CREATE INDEX IF NOT EXISTS idx_assets_tags ON assets USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_assets_updated ON assets(updated_at);
