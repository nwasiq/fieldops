# Architecture

```
fieldops/
├── backend/                 Go API
│   ├── main.go              boot: config, DB, migrations, seed, router
│   ├── internal/api/        Gin handlers + routes.go (HTTP only)
│   ├── internal/services/   business logic + DB queries (+ dbscope.go, ukday.go, crypto.go)
│   ├── internal/models/     GORM structs + constants.go
│   ├── internal/middleware/ auth (JWT), audit registry
│   └── internal/database/   connection, migrations, seed
├── frontend/                React + TS + Vite
│   └── src/{pages,components,lib}   lib/api.ts is the only HTTP client; lib/dateFormat.ts the only date renderer
├── infra/                   Terraform: modules for network, compute (ASG), db (RDS), edge (S3 + CloudFront), plus envs/staging
├── scripts/                 check-*.sh lint scripts (each exits non-zero with the offending lines)
├── docs/                    CONVENTIONS, BUSINESS_RULES, ARCHITECTURE, OPERATIONS, openapi.yaml
├── docker-compose.yml       db (+ optional backend/frontend containers)
├── Makefile                 backend, frontend, test, lint, seed, db-reset
└── .github/workflows/ci.yml go test, vitest, lint scripts, terraform validate
```

## Request flow

Browser → `ApiClient` (`frontend/src/lib/api.ts`, attaches the JWT) → Gin router (`backend/internal/api/routes.go`) → auth middleware (JWT → user + role) → audit middleware (for mutating routes: fetch before-state, run handler, fetch after-state, write `audit_logs`) → handler (parse, call service, write JSON `{ success, data }` or `{ success: false, error }`) → service (all logic and queries, GORM) → PostgreSQL.

## Environment

`.env.example` documents every variable: `DATABASE_URL`, `JWT_SECRET`, `ENCRYPTION_KEY` (32 bytes, base64), `PORT`, `SEED_ON_BOOT`. Nothing reads any other environment variable.
