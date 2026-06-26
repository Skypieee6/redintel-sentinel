-- 0005_audit.up.sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID REFERENCES organizations(id) ON DELETE CASCADE,
    actor_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    action      TEXT NOT NULL,
    target_type TEXT NOT NULL DEFAULT '',
    target_id   TEXT NOT NULL DEFAULT '',
    metadata    JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip          TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_org ON audit_logs(org_id);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at);
