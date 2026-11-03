#!/usr/bin/env bash
set -euo pipefail
if [[ "$GITHUB_EVENT_NAME" != push || "$GITHUB_REF" != refs/heads/* ]]; then
  echo 'Report commits are limited to branch push events.'
  exit 0
fi
cd "$GITHUB_WORKSPACE"
report_dir="$(dirname "$REPORT_JSON")"
git add -f -- "$report_dir/report.json" "$report_dir/report.md" "$report_dir/badge.svg"
if ! git diff --cached --quiet; then
  git -c user.name='github-actions[bot]' -c user.email='41898282+github-actions[bot]@users.noreply.github.com' \
    -c commit.gpgsign=false commit -m 'chore: update C++ Report Card [skip ci]' -- \
    "$report_dir/report.json" "$report_dir/report.md" "$report_dir/badge.svg"
  git push origin "HEAD:$GITHUB_REF_NAME"
fi
