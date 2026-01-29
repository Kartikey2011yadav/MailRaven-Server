#!/bin/bash
# Setup git hooks

HOOK_DIR=".git/hooks"
PRE_COMMIT_SCRIPT="scripts/pre-commit"

if [ ! -d ".git" ]; then
    echo "Error: Not a git repository."
    exit 1
fi

echo "Setting up pre-commit hook..."
cp "$PRE_COMMIT_SCRIPT" "$HOOK_DIR/pre-commit"
chmod +x "$HOOK_DIR/pre-commit"

echo "✅ Hook installed via copy. (Note: On Windows, ensure 'sh' is in your path for git bash)"
