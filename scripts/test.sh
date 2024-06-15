#!/usr/bin/env bash
set -euo pipefail

COVERPROFILE="${COVERPROFILE:-coverage.out}"

# Clear the coverage file
echo "mode: atomic" > "$COVERPROFILE"

gomods=$(find . -name go.mod)

for file in $gomods; do
    printf '=%.0s' {1..80}
    printf "\n===> Testing %s\n" "$file"
    dir=$(dirname "$file")
    pushd "$dir" >/dev/null
    # Create a temporary coverage file for this module
    go test -v -race -coverprofile="tmp.out" -covermode=atomic -timeout 15m ./...
    if [ -f "tmp.out" ]; then
        # Skip the mode line and append to the main coverage file
        tail -n +2 "tmp.out" >> "$COVERPROFILE"
        rm "tmp.out"
    fi
    popd >/dev/null
done