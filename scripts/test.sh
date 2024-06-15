#!/usr/bin/env bash
set -euo pipefail

COVERPROFILE="${COVERPROFILE:-coverage.out}"
WORKSPACE="${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}"
COVERDIR="${COVERDIR:-$WORKSPACE/.coverage}"
mkdir -p "$COVERDIR"

gomods=$(find . -name go.mod)

for file in $gomods; do
    printf '=%.0s' {1..80}
    printf "\n===> Testing %s\n" "$file"
    dir=$(dirname "$file")
    if [[ "$(basename "$dir")" == "." ]]; then
        coverfile="$COVERDIR/root.cover"
    else
        coverfile="$COVERDIR/$(basename "$dir").cover"
    fi
    pushd "$dir" >/dev/null
    go test -v -race -outputdir="$COVERDIR" -coverprofile="$coverfile" -covermode=atomic -timeout 15m ./...
    popd >/dev/null
done