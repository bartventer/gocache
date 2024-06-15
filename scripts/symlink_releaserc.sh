#!/usr/bin/env bash

# This script creates a copy of .releaserc.json in the same directory as go.mod files and adds a "tagFormat" field to it.

set -euo pipefail

echo "(*) Creating copies of .releaserc.json in the same directory as go.mod files and adding a 'tagFormat' field"

# Find all go.mod files
gomods=$(find . -name go.mod)

# create the copy and add the "tagFormat" field
for file in $gomods; do
  dir=$(dirname "$file")
  echo "==> Checking $dir"
  if [[ ! -e $dir/.releaserc.json ]]; then
    # Use the absolute path to .releaserc.json
    target=$(realpath .releaserc.json)
    cp "$target" "$dir"/.releaserc.json
    jq --arg dir "${dir#./}" '. + {tagFormat: "\($dir)/v${version}"}' "$dir"/.releaserc.json >"$dir"/temp.json && mv "$dir"/temp.json "$dir"/.releaserc.json
    echo "  OK. Created copy $dir/.releaserc.json and added 'tagFormat' field"
  else
    echo "  Already exists. Skipping $dir/.releaserc.json"
  fi
done

echo "(*) Done"
