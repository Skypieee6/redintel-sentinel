# RedIntel Sentinel — PRD

## Problem statement
Transform the repository into a production-grade enterprise **Attack Surface
Management (ASM)** platform for **authorized, defensive** security assessments.
No offensive functionality. Deliver the backend foundation.

## Tech stack
Go 1.25 · Gin · Viper · Zap · Cobra · PostgreSQL (pgx/v5) · Redis (go-redis/v9)
· golang-migrate · Docker / docker-compose · GitHub Actions.

## Architecture (implemented)
```
cmd/server        entrypoint -> internal/cli
internal/app      bootstrap & dependency wiring
internal/cli      Cobra commands: serve, migrate, version
internal/config   Viper layered config (defaults -> yaml -> env) + validation
internal/logger   Zap structured logger
internal/version  build metadata via -ldflags
internal/database PostgreSQL pgx pool
internal/cache    Redis client
internal/handlers health, ready, version
internal/middleware request-id, zap logging, recovery
internal/router   Gin engine + routes (/health /ready /version + /api/v1)
internal/server   HTTP server + graceful shutdown
pkg/response      JSON envelope helpers
migrations        golang-migrate SQL (0001 baseline)
configs           default config.yaml
.github/workflows CI pipeline
Dockerfile / docker-compose.yml / Makefile
```

## Implemented (2026-06-26)
- Foundation complete: config, logging, DB+Redis pools, Gin router, middleware,
  graceful shutdown, Cobra CLI, migrations, Docker, CI, Makefile.
- Endpoints verified live: `/health` 200, `/ready` (postgres+redis ok) 200,
  `/version` build info; SIGTERM graceful shutdown confirmed clean.
- `go vet`, `gofmt`, config unit tests all pass. 19 logical commits.

## Backlog (next phases)
- P0: Authentication (JWT) + RBAC; organizations/users schema.
- P1: Projects & asset inventory CRUD; passive asset discovery; HTTP fingerprinting.
- P1: OpenAPI/Swagger docs; rate limiting; structured error catalog.
- P2: Scheduling, notifications, reporting, plugin SDK, distributed workers.
- Tests: handler/integration tests with testcontainers; golangci-lint in CI.

## Notes
- App listens on :8080 by default (configurable via REDINTEL_SERVER_PORT).
- Docker not installed in the build sandbox; Dockerfile/compose validated by review, build to be run in CI / target host.
