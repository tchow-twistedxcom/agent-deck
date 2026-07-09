#!/usr/bin/env bash
# Wrapper for portainer-mcp that reads the API token from the environment.
# Invoked via `op run` so PORTAINER_TOKEN is injected from 1Password at launch.
exec /home/tchow/.claude/bin/portainer-mcp \
  -server twistedx-docker:9443 \
  -token "${PORTAINER_TOKEN}" \
  -disable-version-check \
  "$@"
