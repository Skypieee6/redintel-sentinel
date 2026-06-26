# Test Credentials — RedIntel Sentinel

## Platform Superadmin (auto-seeded on startup)
- Email: `admin@redintel.local`
- Password: `ChangeMe123!`
- Role: superadmin (platform-wide; implicit admin in every org)

Configured via env: `REDINTEL_AUTH_ADMIN_EMAIL`, `REDINTEL_AUTH_ADMIN_PASSWORD`.
Seeding is idempotent (created only if the email does not already exist).

## Auth model
- Access token: JWT (Bearer), default TTL 15m.
- Refresh token: opaque, rotated on use, stored hashed; default TTL 168h.
- API keys: `X-API-Key` header; plaintext shown once on creation.

## Key endpoints (base: `/api/v1`)
- POST `/auth/register`, POST `/auth/login`, POST `/auth/refresh`, POST `/auth/logout`
- GET/PUT `/auth/me`, POST `/auth/change-password`
- POST `/auth/forgot-password`, POST `/auth/reset-password`
- GET/POST `/auth/api-keys`, DELETE `/auth/api-keys/:id`
- GET/POST `/orgs`, GET/PUT/DELETE `/orgs/:orgID`
- `/orgs/:orgID/members`, `/teams`, `/invitations`, `/projects`, `/audit-logs`
- POST `/invitations/accept`
- GET `/admin/users`, GET `/admin/audit-logs` (superadmin)
- Docs: GET `/docs` (Swagger UI), GET `/api/v1/openapi.yaml`

## Local test infra
- PostgreSQL: localhost:5432, db `redintel`, user/pass `postgres`/`postgres`.
- Redis: localhost:6379.
- Run migrations: `go run ./cmd/server migrate up`.
- App default port 8080 (8080 is taken by the platform proxy in this sandbox; run on another port for local smoke tests).
