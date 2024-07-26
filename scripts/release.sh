#!/usr/bin/env bash
set -euo pipefail

# This script is used to release the project in all directories that contain a .releaserc.(json|yml|yaml) file.

if [[ -z "${CI-}" ]] && [[ "$*" != *--dry-run* ]]; then
    echo "🚨 WARNING: You are about to release without the --dry-run flag."
    echo "🚨 WARNING: This will publish the new version(s) and create a new git tag(s)."
    echo "🚨 WARNING: Are you sure you want to continue? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo "🚫 Release process aborted."
        exit 1
    fi
fi

echo "================================================================================"
echo "🔧 Starting Release Process"
echo "================================================================================"

yarn install

releasercs=$(find . \
    -name '.releaserc.json' -type f \
    -o -name '.releaserc.yml' -type f \
    -o -name '.releaserc.yaml' -type f | sort)

for rc in $releasercs; do
    echo "--------------------------------------------------------------------------------"
    echo "Releasing for $rc"
    echo "--------------------------------------------------------------------------------"
    dir=$(dirname "$rc")
    RELEASEDIR=$dir yarn run release "$@"
    echo "  ✔️ Released successfully."
done

echo "================================================================================"
echo "✅ Release Process Completed Successfully"
echo "================================================================================"
