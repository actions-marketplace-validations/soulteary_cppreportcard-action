#!/usr/bin/env bash
set -euo pipefail
# All caller values travel as environment variables and quoted arguments.
python3 - <<'PY'
import os
from pathlib import Path
workspace = Path(os.environ['GITHUB_WORKSPACE']).resolve()
root = (workspace / os.environ['INPUT_DIRECTORY']).resolve()
root.relative_to(workspace)
output = (root / os.environ['INPUT_OUTPUT_DIR']).resolve()
output.relative_to(root)
PY
args=(--directory "$GITHUB_WORKSPACE/$INPUT_DIRECTORY" --mode "$INPUT_MODE"
  --compile-commands "$INPUT_COMPILE_COMMANDS" --output-dir "$INPUT_OUTPUT_DIR"
  --fail-on "$INPUT_FAIL_ON" --threshold "$INPUT_THRESHOLD")
[[ -z "$INPUT_CONFIG" ]] || args+=(--config "$INPUT_CONFIG")
[[ -z "$INPUT_BASELINE" ]] || args+=(--baseline "$INPUT_BASELINE")
result_json="$(mktemp)"
trap 'rm -f "$result_json"' EXIT
export RESULT_JSON="$result_json"
set +e
"$RUNNER_TEMP/cppreportcard-cli" "${args[@]}" --json > "$result_json"
result=$?
set -e
export RESULT_CODE="$result"
python3 - <<'PY'
import json, os
from pathlib import Path
root = (Path(os.environ['GITHUB_WORKSPACE']) / os.environ['INPUT_DIRECTORY']).resolve()
out = (root / os.environ['INPUT_OUTPUT_DIR']).resolve()
p = out / 'report.json'
if not p.is_file():
    raise SystemExit('CLI did not produce a report; inspect the error above')
r = json.loads(Path(os.environ['RESULT_JSON']).read_text())
values = {'status': r['status'], 'grade': r['grade'], 'score': '' if r['score'] is None else r['score'],
          'analysis_complete': str(r['analysis_complete']).lower(), 'new_errors': r['new_errors'],
          'report': out / 'report.md', 'json': p, 'badge': out / 'badge.svg', 'exit_code': os.environ['RESULT_CODE']}
with open(os.environ['GITHUB_OUTPUT'], 'a') as f:
    for key, value in values.items():
        if '\n' in str(value) or '\r' in str(value):
            raise SystemExit('Output contains a newline')
        f.write(f'{key}={value}\n')
with open(os.environ['GITHUB_STEP_SUMMARY'], 'a') as f:
    f.write((out / 'report.md').read_text())
PY
