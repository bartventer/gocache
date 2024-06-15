#!/usr/bin/env bash

# This script creates a symlink to .releaserc.json in the same directory as go.mod files.

set -euo pipefail

echo "(*) Creating symlinks to .releaserc.json in the same directory as go.mod files"

# Find all go.mod files
gomods=$(find . -name go.mod)

# create the symlink
for file in $gomods; do
  dir=$(dirname "$file")
  echo "==> Checking $dir"
  if [[ ! -e $dir/.releaserc.json ]]; then
    # Calculate the relative path to .releaserc.json
    target=$(realpath --relative-to="$dir" .releaserc.json)
    ln -s "$target" "$dir"/.releaserc.json
    echo "  OK. Created symlink $dir/.releaserc.json"
  else
    echo "  Already exists. Skipping $dir/.releaserc.json"
  fi
done

echo "(*) Done"
