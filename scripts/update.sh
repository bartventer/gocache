#!/usr/bin/env bash
set -euo pipefail

echo "================================================================================"
echo "🔧 Starting Dependency Update Process"
echo "================================================================================"

gomods=$(find . -name go.mod -type f -exec dirname {} \; | sort)

for dir in $gomods; do
    echo "--------------------------------------------------------------------------------"
    printf "🔍 Updating dependencies in: %s\n" "$dir"
    echo "--------------------------------------------------------------------------------"
    pushd "$dir" >/dev/null
    go mod tidy
    go get -u ./...
    popd >/dev/null
done

echo "================================================================================"
echo "✅ Dependency Update Process Completed Successfully"
echo "================================================================================"