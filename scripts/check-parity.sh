#!/usr/bin/env bash
# Compare the Go fit engine with the web tool's compass-core.js on fixtures.
# Usage: scripts/check-parity.sh /path/to/University\ Planner
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SOURCE="${1:-/Users/shreyaskali/Startups/Me/Deferred MBA/University Planner}"

if [ ! -f "$SOURCE/compass-core.js" ]; then
  echo "parity: compass-core.js not found in $SOURCE" >&2
  exit 2
fi

status=0
for fixture in "$ROOT"/parity/fixtures/*.json; do
  name="$(basename "$fixture" .json)"
  go_file="$(mktemp)"
  js_file="$(mktemp)"
  (cd "$ROOT" && go run ./cmd/parity-go "$fixture") > "$go_file"
  (cd "$SOURCE" && COMPASS_CORE="$SOURCE/compass-core.js" node "$ROOT/parity/js-engine.mjs" "$fixture") > "$js_file"
  # Compare parsed JSON: key order and HTML escaping (\u0026 vs &) are not
  # semantic differences between the two engines.
  if python3 - "$go_file" "$js_file" <<'PY'
import json, sys
a = json.load(open(sys.argv[1]))
b = json.load(open(sys.argv[2]))
sys.exit(0 if a == b else 1)
PY
  then
    printf 'PASS  %s\n' "$name"
  else
    printf 'FAIL  %s\n' "$name"
    diff -u <(python3 -m json.tool "$go_file") <(python3 -m json.tool "$js_file") | head -30 || true
    status=1
  fi
  rm -f "$go_file" "$js_file"
done
exit "$status"
