# Conventions

These are binding. Some are checked by `make lint` (marked **CI**); the rest are checked in review. A pull request that breaks one is not merged, however good the rest of it is.

## Backend

1. **Three layers, never crossed.** `internal/api` parses HTTP and writes responses. `internal/services` holds every piece of business logic and every database query. `internal/models` holds GORM structs and constants only. A handler that touches `gorm.DB` is a bug.
2. **No hard-coded domain strings. (CI)** Roles, statuses and event kinds come from `internal/models/constants.go` (`models.RoleTechnician`, `models.VisitStatusScheduled`, …). `scripts/check-hardcoded-roles.sh` fails the build on a role literal outside `models`.
3. **Soft deletes are the default.** Every table with `deleted_at` is read through GORM's default scope, so a departed user is invisible to ordinary queries. Anywhere a row is shown for *attribution* (who clocked in, who scheduled a visit) the association must be preloaded unscoped (`services.UnscopedUserAssoc`), otherwise the name renders as "unknown" or, worse, the zero value's fields are read as real data.
4. **Never discard a query error inside a transaction.** A GORM call whose `.Error` is unchecked inside `tx` poisons every later statement. Handle `gorm.ErrRecordNotFound` explicitly where missing is fine; propagate everything else.
5. **Time is stored as instants and read as Europe/London days.** Columns are `timestamptz`. Any "which day is this" question — reports, day buckets, "today" — goes through `services.UKDay` / `services.UKDayStart`, never `time.Now().Truncate(24h)` or host-local `time.Date(...)`. The server's own timezone is irrelevant and must stay irrelevant.
6. **Sensitive columns are encrypted at rest.** `users.licence_number` is stored encrypted; the service layer calls `encryptUserSensitive` before save and `decryptUserSensitive` after read. A raw SQL read of the column returns ciphertext, and that is correct.
7. **Every mutating route has an audit entry.** `PUT`/`POST`/`DELETE` handlers that change a row go through the audit middleware registry in `internal/middleware/audit.go`; a new route is registered there or the request is rejected in review. **(CI)** `scripts/check-audit-coverage.sh` lists mutating routes with no registration.
8. **Pagination is capped and echoed.** List endpoints clamp `page_size` to 100 and return the effective `page_size`. A client that needs a complete list asks for it explicitly; it never asks for a bigger page.

## Frontend

9. **No raw `fetch` in components or pages. (CI)** Every call goes through `ApiClient` in `src/lib/api.ts`. `scripts/check-frontend-fetch.sh` enforces it.
10. **`ApiClient.request()` already unwraps the `{ success, data }` envelope.** Never read `.data` off its result; type the response concretely, never `request<any>`.
11. **Date-filtered lists fetch with server-side date params.** Never filter client-side from a paginated result — rows beyond page 1 silently disappear.
12. **No browser-timezone reads of API timestamps. (CI)** Render through `src/lib/dateFormat.ts` (`formatDateTimeUK`, `ukDatePart`). `scripts/check-naive-timestamp-reads.sh` bans `.split('T')[1]`-style surgery and `toLocale*String` without a `timeZone`.

## Change discipline

13. **Tests first for bug fixes.** A bug fix lands as a test that fails on the current code for the right reason, then the fix. The failing message is the bug's specification.
14. **Docs travel with the change.** A new endpoint updates `docs/openapi.yaml`; a behaviour change updates `docs/BUSINESS_RULES.md`; a test that verifies a documented rule carries `// rule: §X.Y` above it.
15. **Comments say why, not what.** No commit references, ticket numbers or author names in source comments.
16. **Small, reviewable changes.** One concern per pull request. A refactor nobody asked for is a reason to reject, not a bonus.
17. **Nothing secret in the repository.** No credentials, no `.env`, no real customer data. `.env.example` documents every variable.
