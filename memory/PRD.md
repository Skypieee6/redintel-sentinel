# RedIntel Sentinel — PRD

## Problem statement
Production-grade enterprise **Attack Surface Management (ASM)** platform for
**authorized, defensive** security assessments. No offensive functionality.

## Tech stack
Go 1.25 · Gin · Viper · Zap · Cobra · PostgreSQL (pgx/v5) · Redis (go-redis/v9)
· golang-migrate · golang-jwt · bcrypt · Docker / docker-compose · GitHub Actions.

## Architecture
Layered: `cmd/server` -> `internal/cli` -> `internal/app` (wiring) ->
`internal/router` (Gin) -> `internal/middleware` -> `internal/handlers` ->
`internal/service` -> `internal/repository` -> PostgreSQL. Cross-cutting:
`internal/config`, `internal/logger`, `internal/auth`, `internal/models`,
`internal/database`, `internal/cache`, `internal/version`, `pkg/response`.

## Implemented
### Phase 0 — Foundation (2026-06-26)
Config, logging, DB+Redis pools, Gin router, middleware, graceful shutdown,
Cobra CLI (serve/migrate/version), migrations, Docker, CI, Makefile.
Endpoints: `/health`, `/ready`, `/version`.

### Phase 1 — Core platform (2026-06-26)
- Auth: register, login, JWT access + rotating hashed refresh tokens, logout,
  change/reset password, API keys (hashed, shown once), bcrypt, Redis brute-force lockout.
- RBAC: org roles admin/manager/analyst/viewer + platform superadmin (middleware-enforced).
- Organizations: orgs, memberships, teams, team members, email invitations.
- Projects: CRUD with ownership + access control, project members.
- Audit logging: auth/org/project/user events; per-org and platform-wide queries.
- API: REST under `/api/v1`; OpenAPI 3 at `/api/v1/openapi.yaml`; Swagger UI at `/docs`.
- Admin superadmin auto-seeded from env (idempotent).
- Tests: unit (auth, models, config) + full end-to-end HTTP integration test. All pass.

Verification: `go vet` + `gofmt` clean; `go test ./...` green (incl. integration
against live Postgres+Redis); live curl smoke tests for login/org/project/invite/audit/docs.

### Phase 1 — Passive Asset Discovery (2026-06-26)
- Isolated `internal/discovery` package: pluggable `Source` engine doing strictly
  passive, defensive recon — public DNS resolution (A/AAAA/CNAME/MX/NS/TXT),
  Certificate Transparency logs (crt.sh) for subdomains + certificates, and
  reverse DNS over authorized CIDRs. No intrusive scanning. Deduplicates findings.
- Inputs: domain, subdomain, ASN, CIDR. Outputs: subdomains, DNS records, certificates.
- `discovery_jobs` + `discovery_results` tables (migration 0007). Jobs run
  asynchronously with pending→running→completed/failed status and asset counts.
- Findings are persisted as normal `assets` via a new idempotent
  `AssetRepository.Upsert` (ON CONFLICT refresh + attribute merge; reports new vs known).
- RBAC: analyst+ to start, viewer+ to read (reuses project-access rules). Audit
  events: discovery.started / completed / failed.
- API: `POST/GET /orgs/:orgID/projects/:projectID/discovery` (start + history),
  `GET .../discovery/:jobID` (job + findings). OpenAPI updated with schemas.
- Engine injected via `DiscoveryService.SetEngine` for deterministic offline tests.
- Tests: discovery engine unit tests (aggregation, dedupe, partial-failure tolerance,
  normalization) + full integration test (start → poll → assets created → history →
  idempotent re-run). `apitest.TestMain` now applies migrations so CI provisions the
  schema. `gofmt -s` / `go vet` / `go build` / `go test -race` all green from a fresh DB.
- Frontend: Discovery page (start workflow + live-polling history table),
  Discovery detail page (status, stats, results grouped by type, links to assets),
  status badges, error handling. New nav entry + routes. `tsc` + `vite build` clean.

### Phase 5–7 — Asset Inventory, Dashboard & Reporting (2026-06-26)
- Assets: unified `assets` table (domain, subdomain, ip, cidr, asn, dns_record,
  certificate, technology) with JSONB attributes + `text[]` tags. Full CRUD,
  tagging, ILIKE search, type/tag/status filtering and paginated listing
  (returns total). Project-scoped access control (analyst+ to write).
- Project archive/unarchive endpoints.
- Dashboard: total assets, assets-by-type, recent changes, project statistics
  (active/archived + per-project asset counts) and team statistics.
- Reporting: asset-inventory export in JSON, CSV, Markdown and HTML (downloadable).
- OpenAPI updated; unit tests for all report formats; integration test covering
  the full workflow: register → login → org → project → add assets → inventory
  search → dashboard → all 4 report formats → archive. All green.
- Verified live end-to-end via curl (3 assets across types, search, dashboard
  breakdown, all report content-types).

## Backlog (next phases)
- P1: HTTP fingerprinting (defensive); ASN→prefix expansion source for discovery
  (currently ASN seeds resolve via CIDR/reverse-DNS only, no BGP lookup).
- P1: Email delivery for invitations/resets (currently logged); golangci-lint in CI; rate limiting.
- P1: Discovery scheduling (recurring jobs) + graceful drain of in-flight jobs on shutdown.
- P2: Scheduling, notifications, plugin SDK, distributed workers, AI-assisted reporting.

## Notes
- App default port 8080 (8080 occupied by platform proxy in this sandbox; run on
  another port for local smoke tests). Run migrations: `go run ./cmd/server migrate up`.
- The Emergent testing_agent / managed deploy target the standard React/FastAPI/Mongo
  stack; this Go+Postgres+Redis service is validated via its Go integration test suite
  and deploys via Docker/container hosts (see support guidance in chat).
