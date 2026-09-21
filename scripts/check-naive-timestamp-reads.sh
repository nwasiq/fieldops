#!/usr/bin/env bash
# check-naive-timestamp-reads.sh — docs/CONVENTIONS.md §12 (CI)
#
# Fails on browser-timezone reads of API timestamps in frontend/src/**/*.ts and *.tsx,
# excluding src/lib/dateFormat.ts (the one renderer) and test files (*.test.ts[x],
# *.spec.ts[x], __tests__/). Two families:
#
#   1. wall-clock string surgery on an RFC3339 timestamp:
#        .split('T')[1]   .split("T")[1]   .slice(11   .substring(11
#   2. toLocaleString( / toLocaleDateString( / toLocaleTimeString( on a line that does
#      not also contain `timeZone` — without an explicit zone these render in whatever
#      timezone the browser happens to be in.
#
# Output: one file:line:text per offender and exit 1, or one OK line and exit 0.
# Skips with a note (exit 0) while frontend/src does not exist yet.
# REPO_ROOT overrides the repository root so the script can be run against a fixture tree.
set -euo pipefail

REPO_ROOT="${REPO_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"
cd "$REPO_ROOT"

TARGET=frontend/src
Q="'"
SURGERY_PATTERN="\\.split\\([\"$Q]T[\"$Q]\\)\\[1\\]|\\.slice\\(11|\\.substring\\(11"
LOCALE_PATTERN='toLocale(Date|Time)?String\('

if [ ! -d "$TARGET" ]; then
  echo "OK: check-naive-timestamp-reads — $TARGET not present yet, nothing to check"
  exit 0
fi

files=()
while IFS= read -r -d '' f; do files+=("$f"); done < <(
  find "$TARGET" -type f \( -name '*.ts' -o -name '*.tsx' \) \
    ! -path "$TARGET/lib/dateFormat.ts" \
    ! -name '*.test.ts' ! -name '*.test.tsx' ! -name '*.spec.ts' ! -name '*.spec.tsx' \
    ! -path '*/__tests__/*' -print0 | sort -z
)

if [ "${#files[@]}" -eq 0 ]; then
  echo "OK: check-naive-timestamp-reads — no TypeScript files under $TARGET outside lib/dateFormat.ts and tests"
  exit 0
fi

set +e
surgery="$(grep -nHE "$SURGERY_PATTERN" "${files[@]}")"
rc1=$?
locale="$(grep -nHE "$LOCALE_PATTERN" "${files[@]}" | grep -v 'timeZone')"
rc2=$?
set -e
if [ "$rc1" -gt 1 ]; then
  echo "check-naive-timestamp-reads: grep failed (exit $rc1)" >&2
  exit 2
fi

hits=""
[ -n "$surgery" ] && hits="$surgery"
if [ -n "$locale" ]; then
  hits="${hits:+$hits
}$locale"
fi

if [ -n "$hits" ]; then
  echo "$hits"
  count=$(printf '%s\n' "$hits" | wc -l | tr -d ' ')
  echo "FAIL: $count naive timestamp read(s) — render through frontend/src/lib/dateFormat.ts or pass timeZone: 'Europe/London' (docs/CONVENTIONS.md §12)"
  exit 1
fi

echo "OK: check-naive-timestamp-reads — no wall-clock string surgery or zone-less toLocale*String outside lib/dateFormat.ts and tests"
