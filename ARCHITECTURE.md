# Architecture - RedIntel Sentinel

RedIntel Sentinel is an enterprise **Attack Surface Management (ASM)** platform
for **authorized, defensive** security assessments. This document describes the
backend foundation.

> Scope note: The platform is strictly defensive. It is intended to help
> organizations inventory, monitor and assess systems they own or are
> explicitly authorized to test. No offensive/exploitation capability is part
> of this codebase.

## Tech stack

| Concern            | Choice                                  |
|--------------------|-----------------------------------------|
| Language           | Go 1.25                                 |
| HTTP framework     | Gin                                     |
| Configuration      | Viper (+ godotenv for local `.env`)     |
| Logging            | Zap (structured JSON)                   |
| CLI                | Cobra                                   |
| Relational store   | PostgreSQL (pgx/v5 pool)                |
| Cache / ephemeral  | Redis (go-redis/v9)                     |
| Migrations         | golang-migrate                          |
| Packaging          | Multi-stage Docker + docker-compose     |
| CI                 | GitHub Actions                          |

## Layout

```
cmd/server/            Binary entrypoint (delegates to internal/cli)
internal/
  app/                 Application bootstrap & dependency wiring
  cache/               Redis client
  cli/                 Cobra command tree (serve, migrate, version)
  config/              Viper-based layered configuration
  database/            PostgreSQL pgx pool
  handlers/            HTTP handlers (health, ready, version)
  logger/              Zap logger factory
  middleware/          Gin middleware (request id, logging, recovery)
  router/              Gin engine construction & route registration
  server/              HTTP server lifecycle + graceful shutdown
  version/             Build metadata (injected via -ldflags)
pkg/
  response/            Reusable JSON response envelope helpers
configs/               Default config.yaml
migrations/            SQL migrations (golang-migrate)
.github/workflows/     CI pipeline
```

## Configuration model

Layered, later sources win:

1. Built-in defaults (`internal/config`).
2. `config.yaml` discovered in `.`, `./configs`, or `/etc/redintel`.
3. Environment variables prefixed `REDINTEL_` (`.` -> `_`),
   e.g. `REDINTEL_SERVER_PORT=9090`.

A local `.env` file is auto-loaded (best effort) for development.

## Request lifecycle

```
client -> Gin engine
        -> RequestID middleware   (X-Request-ID propagation)
        -> Logger middleware      (structured access logs via Zap)
        -> Recovery middleware    (panic -> 500 + stack log)
        -> handler
```

## Operational endpoints

| Method | Path        | Purpose                                            |
|--------|-------------|----------------------------------------------------|
| GET    | `/health`   | Liveness probe (process up).                       |
| GET    | `/ready`    | Readiness probe (PostgreSQL + Redis reachable).    |
| GET    | `/version`  | Build metadata (version, commit, build time).      |

Aliases `/healthz` and `/readyz` are provided, and the same three endpoints are
mirrored under `/api/v1` for the versioned API surface.

## Lifecycle & graceful shutdown

`serve` traps `SIGINT`/`SIGTERM` via `signal.NotifyContext`. On signal, the HTTP
server stops accepting new connections and drains in-flight requests within
`server.shutdown_timeout`, after which datastore connections are closed.

## Roadmap

The foundation deliberately exposes an empty `/api/v1` group. Subsequent phases
add feature modules (authentication, projects, asset inventory, passive
discovery, reporting) as self-contained packages registering their routes onto
that group.
