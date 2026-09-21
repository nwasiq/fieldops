# fieldops

A small field-service scheduling system: dispatchers schedule **visits** by **technicians** to customer **sites**; technicians clock in and out of visits from a phone; managers read a daily report. It is deliberately shaped like a real production system — conventions, an audit trail, encrypted-at-rest data, timezone discipline, CI gates and infrastructure as code — because it exists to be maintained, not to be admired.

It is the working material for a maintenance-engineer assessment. The assessment tasks are handed to a candidate separately; this repository is the system they work on.

## Stack

| Part | Tech | Port |
|---|---|---|
| `backend/` | Go 1.22+, Gin, GORM, PostgreSQL | 8080 |
| `frontend/` | React 18, TypeScript, Vite, Vitest | 3000 |
| `infra/` | Terraform (AWS: VPC, ASG, RDS, S3, CloudFront) — plans offline, never applied from this repo | — |
| `scripts/` | CI lint scripts run by `.github/workflows/ci.yml` and `make lint` | — |

## Quick start

```bash
cp .env.example .env
docker compose up -d db          # PostgreSQL 16 on localhost:5432
make backend                     # migrates + seeds, serves :8080
make frontend                    # serves :3000, proxies /api to :8080
make test                        # go test ./... + vitest run
make lint                        # every CI lint script
```

Seed logins: `admin@fieldops.local`, `dispatcher@fieldops.local`, `tech1@fieldops.local` … `tech3@fieldops.local`, all with password `fieldops`.

## Read before changing anything

- `docs/CONVENTIONS.md` — the rules every change must follow. CI enforces some of them; reviewers enforce the rest.
- `docs/BUSINESS_RULES.md` — what the system is supposed to do, with numbered rules that tests cite.
- `docs/ARCHITECTURE.md` — where things live and how a request flows.
- `docs/OPERATIONS.md` — how it is deployed, where logs are, how to restore the database.
