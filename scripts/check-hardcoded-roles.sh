#!/usr/bin/env bash
# check-hardcoded-roles.sh — docs/CONVENTIONS.md §2 (CI)
#
# Fails on a role string literal — "admin", "dispatcher" or "technician" — in any
# backend/internal/**/*.go file other than internal/models/ (where the constants live)
# and *_test.go. Roles come from models.Role* constants; a literal is a typo waiting to
# pass code review.
#
# Output: one file:line:text per offender and exit 1, or one OK line and exit 0.
# Skips with a note (exit 0) while backend/internal does not exist yet.
# REPO_ROOT overrides the repository root so the script can be run against a fixture tree.
set -euo pipefail

REPO_ROOT="${REPO_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"
cd "$REPO_ROOT"

TARGET=backend/internal
PATTERN='"(admin|dispatcher|technician)"'

if [ ! -d "$TARGET" ]; then
  echo "OK: check-hardcoded-roles — $TARGET not present yet, nothing to check"
  exit 0
fi

files=()
while IFS= read -r -d '' f; do files+=("$f"); done < <(
  find "$TARGET" -type f -name '*.go' ! -name '*_test.go' ! -path "$TARGET/models/*" -print0 | sort -z
)

if [ "${#files[@]}" -eq 0 ]; then
  echo "OK: check-hardcoded-roles — no Go files under $TARGET outside models/ and tests"
  exit 0
fi

set +e
# A struct tag such as `json:"technician"` names a field, not a role, so lines
# carrying a json tag are not literals to flag.
hits="$(grep -nHE "$PATTERN" "${files[@]}" | grep -v 'json:"' || true)"
rc=$?
set -e
if [ "$rc" -gt 1 ]; then
  echo "check-hardcoded-roles: grep failed (exit $rc)" >&2
  exit 2
fi

if [ -n "$hits" ]; then
  echo "$hits"
  count=$(printf '%s\n' "$hits" | wc -l | tr -d ' ')
  echo "FAIL: $count hard-coded role literal(s) — use models.Role* from backend/internal/models/constants.go (docs/CONVENTIONS.md §2)"
  exit 1
fi

echo "OK: check-hardcoded-roles — no role literals outside backend/internal/models and tests"
