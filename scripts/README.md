# scripts

CI lint scripts. `make lint` and the `lint` job in `.github/workflows/ci.yml` run every `check-*.sh`
in turn and stop at the first failure.

Each script:

- prints one `file:line:text` line per offender and exits 1, or one `OK:` line and exits 0;
- skips with a note (exit 0) while the directory it checks does not exist yet, so the lint job is
  green on a partial checkout and turns real the moment `backend/` or `frontend/` lands;
- honours `REPO_ROOT=<dir>` to run against a fixture tree instead of the repository — that is how
  the scripts are proven to fail (plant a violation, run, remove it, run again);
- runs under bash 3.2 (macOS default) and bash 5 (CI), with BSD or GNU grep/find/sed.

| Script | Convention | What it flags |
|---|---|---|
| `check-hardcoded-roles.sh` | §2 | `"admin"` / `"dispatcher"` / `"technician"` literals in `backend/internal/**/*.go` outside `internal/models/` and `*_test.go` |
| `check-frontend-fetch.sh` | §9 | raw `fetch(` in `frontend/src/**/*.ts{,x}` outside `src/lib/api.ts` and tests |
| `check-naive-timestamp-reads.sh` | §12 | `.split('T')[1]`, `.slice(11`, `.substring(11`, and `toLocale*String(` without `timeZone` in `frontend/src/**/*.ts{,x}` outside `src/lib/dateFormat.ts` and tests |
| `check-audit-coverage.sh` | §7 | a `POST`/`PUT`/`PATCH`/`DELETE` route in `backend/internal/api/routes.go` with no `registry.Register(...)` in `backend/internal/middleware/audit.go`; the inline `ALLOWLIST` (currently `POST /api/auth/login`) carries documented exceptions |

`check-audit-coverage.sh` resolves Gin group prefixes by name from `x := parent.Group("/p")` lines
and only sees routes written as `ident.METHOD("<literal path>", …)`; the exact supported shapes are
in its header. A mutating route written any other way is invisible to it and must not be added.
