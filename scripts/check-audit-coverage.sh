#!/usr/bin/env bash
# check-audit-coverage.sh — docs/CONVENTIONS.md §7 (CI)
#
# Every mutating route in backend/internal/api/routes.go must have an audit registration
# in backend/internal/middleware/audit.go. A route with none is listed and the script
# exits 1; the inline ALLOWLIST below carries the documented exceptions.
#
# routes.go — the line shapes this script understands (leading whitespace allowed):
#   api := r.Group("/api")                group declaration: prefix(api) = prefix(r) + "/api"
#   visits := api.Group("/visits", mw)    nested group: parent resolved by name from an earlier line
#   visits.POST("", h.Create)             route: full path = prefix(visits) + ""
#   api.DELETE("/visits/:id", h.Delete)   METHOD is one of POST PUT PATCH DELETE
# An identifier never declared with .Group( (r, router, engine) has prefix "". Declarations
# are read in file order and the most recent one for a name wins, so resolution is by name,
# not by Go scope. Invisible to this script, and therefore not to be used for mutating routes:
# r.Handle("POST", ...), r.Any(...), paths built from constants or fmt.Sprintf, chained
# r.Group("/x").POST(...), routes registered outside routes.go.
#
# audit.go — the line shapes this script understands:
#   registry.Register("POST", "/api/visits", ...)
#   registry.Register(http.MethodPost, "/api/visits", ...)
# (any receiver before .Register; method as a string literal or a net/http constant).
#
# Paths are compared literally after collapsing "//" — parameter names must match exactly.
#
# Output: one file:line:text per unregistered route (with the resolved METHOD /path) and
# exit 1, or one OK line and exit 0. Skips with a note (exit 0) while routes.go does not
# exist yet. REPO_ROOT overrides the repository root for running against a fixture tree.
set -euo pipefail

REPO_ROOT="${REPO_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"
cd "$REPO_ROOT"

ROUTES=backend/internal/api/routes.go
AUDIT=backend/internal/middleware/audit.go

# "METHOD /full/path|reason" — a mutating route that legitimately has no audit entry.
ALLOWLIST=(
  "POST /api/auth/login|verifies a password and issues a token; no row changes, so there is no before/after state to record"
)

if [ ! -f "$ROUTES" ]; then
  echo "OK: check-audit-coverage — $ROUTES not present yet, nothing to check"
  exit 0
fi

re_group='^[[:space:]]*([A-Za-z_][A-Za-z0-9_]*)[[:space:]]*:?=[[:space:]]*([A-Za-z_][A-Za-z0-9_]*)\.Group\("([^"]*)"'
re_route='^[[:space:]]*([A-Za-z_][A-Za-z0-9_]*)\.(POST|PUT|PATCH|DELETE)\("([^"]*)"'
re_register='\.Register\(("(POST|PUT|PATCH|DELETE)"|http\.Method(Post|Put|Patch|Delete)),[[:space:]]*"([^"]*)"'

# --- registrations from audit.go ------------------------------------------

registered=""
if [ -f "$AUDIT" ]; then
  while IFS= read -r line || [ -n "$line" ]; do
    if [[ $line =~ $re_register ]]; then
      method="${BASH_REMATCH[2]}"
      if [ -z "$method" ]; then
        case "${BASH_REMATCH[3]}" in
          Post) method=POST ;; Put) method=PUT ;; Patch) method=PATCH ;; Delete) method=DELETE ;;
        esac
      fi
      path="$(printf '%s' "${BASH_REMATCH[4]}" | sed -e 's#//*#/#g')"
      registered="$registered
$method $path"
    fi
  done < "$AUDIT"
else
  echo "note: $AUDIT not present — every mutating route counts as unregistered"
fi

is_registered() {
  printf '%s\n' "$registered" | grep -qxF "$1"
}

is_allowlisted() {
  local entry
  for entry in "${ALLOWLIST[@]}"; do
    [ "${entry%%|*}" = "$1" ] && return 0
  done
  return 1
}

# --- routes from routes.go ------------------------------------------------

group_names=()
group_prefixes=()

prefix_of() {
  local i
  for ((i = ${#group_names[@]} - 1; i >= 0; i--)); do
    if [ "${group_names[$i]}" = "$1" ]; then
      printf '%s' "${group_prefixes[$i]}"
      return
    fi
  done
  printf ''
}

failures=""
lineno=0
while IFS= read -r line || [ -n "$line" ]; do
  lineno=$((lineno + 1))
  if [[ $line =~ $re_group ]]; then
    group_names+=("${BASH_REMATCH[1]}")
    group_prefixes+=("$(prefix_of "${BASH_REMATCH[2]}")${BASH_REMATCH[3]}")
    continue
  fi
  if [[ $line =~ $re_route ]]; then
    method="${BASH_REMATCH[2]}"
    path="$(printf '%s' "$(prefix_of "${BASH_REMATCH[1]}")${BASH_REMATCH[3]}" | sed -e 's#//*#/#g')"
    route="$method $path"
    if is_registered "$route" || is_allowlisted "$route"; then
      continue
    fi
    text="$(printf '%s' "$line" | sed -e 's/^[[:space:]]*//')"
    failures="${failures:+$failures
}$ROUTES:$lineno:$text  # $route has no Register entry in $AUDIT"
  fi
done < "$ROUTES"

if [ -n "$failures" ]; then
  echo "$failures"
  count=$(printf '%s\n' "$failures" | wc -l | tr -d ' ')
  echo "FAIL: $count mutating route(s) without an audit registration — add registry.Register(...) in $AUDIT or, if no row changes, an ALLOWLIST entry with the reason (docs/CONVENTIONS.md §7)"
  exit 1
fi

echo "OK: check-audit-coverage — every mutating route in $ROUTES is registered in $AUDIT or allowlisted"
