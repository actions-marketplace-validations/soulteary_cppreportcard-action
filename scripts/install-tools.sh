#!/usr/bin/env bash
set -euo pipefail
[[ "$RUNNER_OS" == Linux ]] || { echo 'Automatic installation supports Ubuntu Linux only; set install_tools=false and provide tools.' >&2; exit 2; }
[[ "$LLVM_VERSION" =~ ^[0-9]+$ ]] || { echo 'Invalid LLVM major version' >&2; exit 2; }
sudo apt-get update -qq
sudo apt-get install -y --no-install-recommends "clang-format-$LLVM_VERSION" "clang-tidy-$LLVM_VERSION" cppcheck
bin="$RUNNER_TEMP/cppreportcard-tools"
mkdir -p "$bin"
ln -sf "$(command -v "clang-format-$LLVM_VERSION")" "$bin/clang-format"
ln -sf "$(command -v "clang-tidy-$LLVM_VERSION")" "$bin/clang-tidy"
printf '%s\n' "$bin" >> "$GITHUB_PATH"
