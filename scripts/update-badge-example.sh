#!/usr/bin/env bash
# Generate the README preview from real analysis of the bundled CMake fixture.
set -euo pipefail
repo="$(cd "$(dirname "$0")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cp "$repo/testdata/cmake/CMakeLists.txt" "$repo/testdata/cmake/clean.cpp" "$repo/testdata/cmake/.clang-format" "$tmp/"
cmake -S "$tmp" -B "$tmp/build" -G 'Unix Makefiles'
cmake --build "$tmp/build"
cd "$repo"
go run ./cmd/cppreportcard-cli --directory "$tmp" --mode full \
  --output-dir "$repo/.github/cppreportcard-example" --fail-on findings
