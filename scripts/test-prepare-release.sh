#!/usr/bin/env bash
set -euo pipefail
repo="$(cd "$(dirname "$0")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/dist"
export GH_CALL_LOG="$tmp/calls"
cat > "$tmp/bin/gh" <<'MOCK'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$GH_CALL_LOG"
if [[ "$1 $2" == 'release view' ]]; then
  case "$MOCK_STATE" in
    missing) exit 1 ;;
    draft) printf 'true\n' ;;
    published) printf 'false\n' ;;
  esac
fi
MOCK
chmod +x "$tmp/bin/gh"
export PATH="$tmp/bin:$PATH"
cd "$tmp"
for n in 1 2 3 4 5 6; do touch "dist/$n.tar.gz"; done
touch dist/checksums.txt
export MOCK_STATE=missing
bash "$repo/scripts/prepare-release.sh" v1.0.1
grep -q 'release create .*--draft' "$GH_CALL_LOG"
: > "$GH_CALL_LOG"
export MOCK_STATE=draft
bash "$repo/scripts/prepare-release.sh" v1.0.1
grep -q 'release upload .*--clobber' "$GH_CALL_LOG"
: > "$GH_CALL_LOG"
export MOCK_STATE=published
if bash "$repo/scripts/prepare-release.sh" v1.0.1; then
  echo 'Published release was not protected' >&2; exit 1
fi
if grep -q 'release upload\|release create' "$GH_CALL_LOG"; then
  echo 'Published release was mutated' >&2; exit 1
fi
if bash "$repo/scripts/prepare-release.sh" invalid; then
  echo 'Invalid version was accepted' >&2; exit 1
fi
