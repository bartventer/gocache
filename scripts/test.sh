#!/usr/bin/env bash
set -euo pipefail

COVERPROFILE="${COVERPROFILE:-coverage.out}"
WORKSPACE="${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}"
COVERDIR="${COVERDIR:-$WORKSPACE/.coverage}"
mkdir -p "$COVERDIR"

gomods=$(find . -name go.mod -type f -exec dirname {} \; | sort)

for dir in $gomods; do
    printf '\n\n%s\n' "$(printf '=%.0s' {1..80})"
    printf "🐛 Testing module at path: %s\n" "$dir"
    printf '%s\n' "$(printf '=%.0s' {1..80})"
    if [[ "$(basename "$dir")" == "." ]]; then
        coverfile="$COVERDIR/root.cover"
    else
        coverfile="$COVERDIR/$(basename "$dir").cover"
    fi
    pushd "$dir" >/dev/null
    go test -v -race -outputdir="$COVERDIR" -coverprofile="$coverfile" -covermode=atomic -timeout 15m ./...
    popd >/dev/null
done