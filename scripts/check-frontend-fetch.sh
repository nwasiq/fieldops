#!/usr/bin/env bash
# check-frontend-fetch.sh — docs/CONVENTIONS.md §9 (CI)
#
# Fails on a raw fetch( call in any frontend/src/**/*.ts or *.tsx file other than
# src/lib/api.ts (the one HTTP client) and test files (*.test.ts[x], *.spec.ts[x],
# anything under a __tests__/ directory). Every API call goes through ApiClient.
#
# The match is `fetch(` not preceded by an identifier character, so window.fetch( and
# globalThis.fetch( are caught while refetch( / prefetch( are not.
#
# Output: one file:line:text per offender and exit 1, or one OK line and exit 0.
# Skips with a note (exit 0) while frontend/src does not exist yet.
# REPO_ROOT overrides the repository root so the script can be run against a fixture tree.
set -euo pipefail

REPO_ROOT="${REPO_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"
cd "$REPO_ROOT"

TARGET=frontend/src
PATTERN='(^|[^A-Za-z0-9_])fetch\('

if [ ! -d "$TARGET" ]; then
  echo "OK: check-frontend-fetch — $TARGET not present yet, nothing to check"
  exit 0
fi

files=()
while IFS= read -r -d '' f; do files+=("$f"); done < <(
  find "$TARGET" -type f \( -name '*.ts' -o -name '*.tsx' \) \
    ! -path "$TARGET/lib/api.ts" \
    ! -name '*.test.ts' ! -name '*.test.tsx' ! -name '*.spec.ts' ! -name '*.spec.tsx' \
    ! -path '*/__tests__/*' -print0 | sort -z
)

if [ "${#files[@]}" -eq 0 ]; then
  echo "OK: check-frontend-fetch — no TypeScript files under $TARGET outside lib/api.ts and tests"
  exit 0
fi

set +e
hits="$(grep -nHE "$PATTERN" "${files[@]}")"
rc=$?
set -e
if [ "$rc" -gt 1 ]; then
  echo "check-frontend-fetch: grep failed (exit $rc)" >&2
  exit 2
fi

if [ -n "$hits" ]; then
  echo "$hits"
  count=$(printf '%s\n' "$hits" | wc -l | tr -d ' ')
  echo "FAIL: $count raw fetch( call(s) — go through ApiClient in frontend/src/lib/api.ts (docs/CONVENTIONS.md §9)"
  exit 1
fi

echo "OK: check-frontend-fetch — no raw fetch( outside frontend/src/lib/api.ts and tests"
