#!/usr/bin/env bash
# Install Claude Code config from this repo via symlinks.
# Run once on a new machine after cloning, or after changing which files to track.
# Requires: op CLI authenticated (OP_SERVICE_ACCOUNT_TOKEN set)

set -euo pipefail
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_DIR="$REPO_DIR/config"

backup() {
  local target="$1"
  if [[ -e "$target" && ! -L "$target" ]]; then
    mv "$target" "${target}.bak.$(date +%Y%m%d%H%M%S)"
    echo "  backed up: $target"
  fi
}

echo "Installing Claude Code config from $CONFIG_DIR ..."

# ~/.mcp.json
backup "$HOME/.mcp.json"
ln -sf "$CONFIG_DIR/mcp.json" "$HOME/.mcp.json"
echo "  linked: ~/.mcp.json"

# ~/.claude/settings.local.json
mkdir -p "$HOME/.claude"
backup "$HOME/.claude/settings.local.json"
ln -sf "$CONFIG_DIR/claude/settings.local.json" "$HOME/.claude/settings.local.json"
echo "  linked: ~/.claude/settings.local.json"

# ~/.claude/CLAUDE.md
backup "$HOME/.claude/CLAUDE.md"
ln -sf "$CONFIG_DIR/claude/CLAUDE.md" "$HOME/.claude/CLAUDE.md"
echo "  linked: ~/.claude/CLAUDE.md"

echo ""
echo "Done. Verify op can resolve the token references:"
echo "  op run -- env | grep -E 'PORTAINER|SLACK'"
