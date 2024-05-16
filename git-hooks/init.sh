#!/bin/bash

# Get the current script's directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Get the .git/hooks directory
HOOKS_DIR="$(git rev-parse --show-toplevel)/.git/hooks"

# Loop through all files in the git-hooks directory
for file in "$SCRIPT_DIR"/*; do
    # Exclude the current file
    if [[ "$file" != "$SCRIPT_DIR/init.sh" ]]; then
        # Create a symbolic link in the .git/hooks directory
        ln -s "$file" "$HOOKS_DIR/$(basename "$file")"
    fi
done
