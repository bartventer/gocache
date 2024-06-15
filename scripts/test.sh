#!/usr/bin/env bash
set -euo pipefail

COVERPROFILE="${COVERPROFILE:-coverage.out}"
WORKSPACE="${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}"

# Clear the coverage file
echo "mode: atomic" > "$COVERPROFILE"

gomods=$(find . -name go.mod)

for file in $gomods; do
    printf '=%.0s' {1..80}
    printf "\n===> Testing %s\n" "$file"
    dir=$(dirname "$file")
    pushd "$dir" >/dev/null
    # Create a temporary coverage file for this module
    tmpfile=$(mktemp --suffix=".out" --tmpdir="$WORKSPACE")
    go test -v -race -outputdir="$WORKSPACE" -coverprofile="$tmpfile" -covermode=atomic -timeout 15m ./...
    if [[ -f "$tmpfile" ]]; then
        # Skip the mode line and append to the main coverage file
        tail -n +2 "$tmpfile" >> "$COVERPROFILE"
        rm "$tmpfile"
    fi
    popd >/dev/null
done