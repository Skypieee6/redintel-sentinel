-- 0007_discovery.up.sql
-- Passive asset discovery: jobs and their findings. Findings are also persisted
-- as normal rows in the assets table; discovery_results links a job to those
-- assets for history and auditing.
CREATE TABLE IF NOT EXISTS discovery_jobs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id     UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    input_type     TEXT NOT NULL CHECK (input_type IN ('domain','subdomain','asn','cidr')),
    input_value    TEXT NOT NULL,
    sources        TEXT[] NOT NULL DEFAULT '{}',
    status         TEXT NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending','running','completed','failed')),
    error          TEXT NOT NULL DEFAULT '',
    assets_found   INTEGER NOT NULL DEFAULT 0,
    assets_created INTEGER NOT NULL DEFAULT 0,
    created_by     UUID REFERENCES users(id) ON DELETE SET NULL,
    started_at     TIMESTAMPTZ,
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_discovery_jobs_org ON discovery_jobs(org_id);
CREATE INDEX IF NOT EXISTS idx_discovery_jobs_project ON discovery_jobs(project_id);
CREATE INDEX IF NOT EXISTS idx_discovery_jobs_status ON discovery_jobs(status);
CREATE INDEX IF NOT EXISTS idx_discovery_jobs_created ON discovery_jobs(created_at);

CREATE TABLE IF NOT EXISTS discovery_results (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id     UUID NOT NULL REFERENCES discovery_jobs(id) ON DELETE CASCADE,
    asset_id   UUID REFERENCES assets(id) ON DELETE SET NULL,
    type       TEXT NOT NULL,
    value      TEXT NOT NULL,
    source     TEXT NOT NULL DEFAULT '',
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_new     BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_discovery_results_job ON discovery_results(job_id);
CREATE INDEX IF NOT EXISTS idx_discovery_results_asset ON discovery_results(asset_id);
