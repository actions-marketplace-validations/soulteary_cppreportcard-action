#!/usr/bin/env bash
set -euo pipefail
version="${1:?usage: build-release.sh v1.0.0}"
[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][A-Za-z0-9.-]+)?$ ]] || exit 2
mkdir -p dist
for os in linux darwin windows; do
  for arch in amd64 arm64; do
    name="cppreportcard-cli"; [[ "$os" != windows ]] || name+='.exe'
    dir="$(mktemp -d)"
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags "-s -w -X main.version=$version" -o "$dir/$name" ./cmd/cppreportcard-cli
    cp LICENSE "$dir/"
    tar -czf "dist/cppreportcard_${version}_${os}_${arch}.tar.gz" -C "$dir" "$name" LICENSE
    rm -rf "$dir"
  done
done
(cd dist && sha256sum ./*.tar.gz > checksums.txt)
