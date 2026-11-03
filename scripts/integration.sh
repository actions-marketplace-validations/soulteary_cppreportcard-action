#!/usr/bin/env bash
set -euo pipefail
repo="$(cd "$(dirname "$0")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cp -R "$repo/testdata/cmake/." "$tmp/"
cmake -S "$tmp" -B "$tmp/build" -G 'Unix Makefiles'
cmake --build "$tmp/build"
go build -o "$tmp/cli" "$repo/cmd/cppreportcard-cli"
"$tmp/cli" --directory "$tmp" --mode full --fail-on findings
python3 - "$tmp" <<'PY'
import json, sys
from pathlib import Path
p = Path(sys.argv[1]); r = json.loads((p/'.cppreportcard/report.json').read_text())
assert r['analysis_complete'] and r['grade'] == 'A+', r
assert all(c['analyzed'] == c['total'] for c in r['checks']), r
PY
# Defect detected; reports still emitted; baseline suppresses only new-error gate.
cat > "$tmp/clean.cpp" <<'CPP'
int main() {
  int *p = nullptr;
  return *p;
}
CPP
set +e
"$tmp/cli" --directory "$tmp" --mode full --fail-on errors
code=$?
set -e
[[ "$code" == 1 ]]
cp "$tmp/.cppreportcard/report.json" "$tmp/baseline.json"
"$tmp/cli" --directory "$tmp" --mode full --baseline baseline.json --fail-on new-errors
# Missing generated/dependency header is incomplete, never A+.
printf '#include "missing-generated-header.h"\nint main() {}\n' > "$tmp/clean.cpp"
set +e
"$tmp/cli" --directory "$tmp" --mode full --fail-on none
code=$?
set -e
[[ "$code" == 2 ]]
python3 - "$tmp/.cppreportcard/report.json" <<'PY'
import json, sys
r = json.load(open(sys.argv[1]))
assert not r['analysis_complete'] and r['score'] is None and r['grade'] == 'incomplete', r
PY
