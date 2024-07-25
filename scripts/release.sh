#!/usr/bin/env bash
set -euo pipefail

# This script is used to release the project in all directories that contain a .releaserc.json file.

echo "================================================================================"
echo "🔧 Starting Release Process"
echo "================================================================================"

yarn install

releasedirs=$(find . -name '.releaserc.json' -type f -exec dirname {} \; | sort)

for dir in $releasedirs; do
    echo "--------------------------------------------------------------------------------"
    echo "Releasing in: $dir"
    echo "--------------------------------------------------------------------------------"
    GOMODDIR=$dir yarn run release "$@"
    echo "  ✔️ Released successfully."
done

echo "================================================================================"
echo "✅ Release Process Completed Successfully"
echo "================================================================================"
