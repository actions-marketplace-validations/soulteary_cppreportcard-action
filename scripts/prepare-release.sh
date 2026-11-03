#!/usr/bin/env bash
# Keep publication as a UI step so maintainers can select Marketplace categories.
set -euo pipefail
version="${1:?usage: prepare-release.sh v1.0.1}"
[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][A-Za-z0-9.-]+)?$ ]] || { echo 'Invalid release version' >&2; exit 2; }
shopt -s nullglob
assets=(dist/*.tar.gz dist/checksums.txt)
[[ ${#assets[@]} -eq 7 && -f dist/checksums.txt ]] || { echo 'Expected six CLI archives and checksums.txt' >&2; exit 2; }
if state="$(gh release view "$version" --json isDraft --jq .isDraft 2>/dev/null)"; then
  [[ "$state" == true ]] || { echo 'Refusing to change assets on a published release' >&2; exit 2; }
  gh release upload "$version" "${assets[@]}" --clobber
else
  gh release create "$version" "${assets[@]}" --verify-tag --generate-notes \
    --draft --title "C++ Report Card $version"
fi
if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  cat >> "$GITHUB_STEP_SUMMARY" <<SUMMARY
## Marketplace release draft ready

CLI archives and SHA-256 checksums have been attached to the draft for $version.

1. Open the repository's Releases page and edit this draft.
2. Select **Publish this Action to the GitHub Marketplace**.
3. Select **Code quality** as the primary category (and **Code review** if appropriate).
4. Publish the release, then promote the stable major tag following .github/MARKETPLACE.md.

GitHub checks the listing name and account eligibility in the release form.
SUMMARY
fi
