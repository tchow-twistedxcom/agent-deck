# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Agent Deck is a terminal session manager for AI coding agents (Claude Code, Gemini CLI, OpenCode, Codex, Cursor). It provides a TUI built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [tmux](https://github.com/tmux/tmux) for managing multiple AI agent sessions with features like:

- Session forking (duplicate Claude conversations with full context)
- MCP server management (toggle MCP servers per-project without editing config files)
- MCP Socket Pool (share MCP processes across sessions to reduce memory)
- Smart status detection (running/waiting/idle/error based on terminal content)
- Fuzzy search across all sessions

## Development Commands

```bash
# Build and run
make build          # Build binary to ./build/agent-deck
make run            # Run directly with go run
make dev            # Development with auto-reload (requires 'air')
make install        # Install to /usr/local/bin (requires sudo)
make install-user   # Install to ~/.local/bin (no sudo)

# Quality
make test           # Run all tests
make lint           # Run golangci-lint
make fmt            # Format code with go fmt

# Release
make release        # Cross-compile for darwin/linux (amd64/arm64)
make clean          # Remove build artifacts
```

### Debug Mode

```bash
AGENTDECK_DEBUG=1 agent-deck    # Enables debug logging to ~/.agent-deck/debug.log
```

## Architecture

### Package Structure

```
cmd/agent-deck/     # CLI entry point and subcommand handlers
├── main.go         # TUI launcher, global flags, CLI routing
├── session_cmd.go  # Session subcommands (start/stop/restart/fork/attach/show)
├── mcp_cmd.go      # MCP subcommands (list/attach/detach/attached)
└── group_cmd.go    # Group subcommands (list/create/delete/move)

internal/
├── session/        # Core domain logic
│   ├── instance.go     # Session instance (status, Claude/Gemini integration)
│   ├── storage.go      # JSON persistence (~/.agent-deck/profiles/<profile>/sessions.json)
│   ├── groups.go       # Hierarchical group tree
│   ├── claude.go       # Claude-specific: session ID detection, forking
│   ├── gemini.go       # Gemini-specific: session handling
│   ├── mcp_catalog.go  # MCP config parsing from config.toml
│   ├── pool_manager.go # MCP socket pool lifecycle
│   └── userconfig.go   # User config (~/.agent-deck/config.toml)
│
├── tmux/           # tmux abstraction layer
│   ├── tmux.go     # Session CRUD, cache management, subprocess optimization
│   ├── detector.go # Status detection (content hashing, prompt patterns)
│   └── watcher.go  # File system watching for activity
│
├── mcppool/        # MCP socket pooling
│   ├── pool_simple.go  # Pool manager for shared MCP processes
│   └── socket_proxy.go # Unix socket proxy for MCP communication
│
├── ui/             # Bubble Tea TUI components
│   ├── home.go         # Main view, status polling, key handling
│   ├── list.go         # Session list rendering
│   ├── tree.go         # Collapsible group tree
│   ├── mcp_dialog.go   # MCP manager modal
│   ├── forkdialog.go   # Fork session modal
│   └── styles.go       # lipgloss styling
│
├── update/         # Self-update mechanism
│   └── update.go   # GitHub release checking and binary replacement
│
└── profile/        # Profile auto-detection
    └── detect.go   # Detect profile from tmux session name
```

### Key Concepts

**Session Instance** (`internal/session/instance.go`):
- Wraps a tmux session with tool-specific behavior (Claude/Gemini/shell)
- Tracks status via content hashing (not process state)
- For Claude: captures session ID from tmux environment, enables fork/resume
- For Gemini: captures session ID for resume functionality

**Status Detection** (`internal/tmux/detector.go`):
- Uses content hashing + prompt pattern matching
- Status: `running` (content changing), `waiting` (stable with input prompt), `idle` (shell prompt), `error` (session doesn't exist)
- Session cache reduces subprocess spawns from O(n) to O(1) per tick

**MCP Management**:
- Available MCPs defined in `~/.agent-deck/config.toml` under `[mcps.*]`
- Per-project MCPs written to `<project>/.mcp.json`
- Global MCPs stored in Claude's config (`~/.claude/settings.json`)
- Socket pool shares MCP processes across sessions via Unix sockets

**Profile System**:
- Multiple profiles for organizing sessions (work/personal/etc)
- Each profile has independent sessions.json
- Data stored in `~/.agent-deck/profiles/<name>/`

### Data Flow

1. **TUI Startup**: `main.go` → `ui.NewHomeWithProfile()` → loads sessions from storage
2. **Status Polling**: Every 500ms, `tmux.RefreshSessionCache()` → update all session statuses
3. **Session Start**: `Instance.Start()` → builds tool-specific command → `tmux.Session.Start()`
4. **MCP Toggle**: `mcp_dialog.go` → `WriteMCPJsonFromConfig()` → writes `.mcp.json` → restart session

### Claude Integration Pattern

The "capture-resume" pattern ensures session ID is always known:
1. Start Claude in print mode (`claude -p "." --output-format json`)
2. Extract session_id from JSON output
3. Store in tmux environment (`tmux set-environment CLAUDE_SESSION_ID`)
4. Resume session interactively (`claude --resume <id>`)

This enables fork and restart features even after agent-deck restarts.

## Testing

```bash
go test -v ./...                    # All tests
go test -v ./internal/session/...   # Package-specific
go test -run TestFoo ./...          # Single test
```

Tests use `testmain_test.go` files to set up test fixtures. Some tests require tmux to be running.

## Configuration Files

- `~/.agent-deck/config.toml` - User configuration (MCPs, profiles, updates)
- `~/.agent-deck/profiles/<profile>/sessions.json` - Session data
- `<project>/.mcp.json` - Per-project MCP configuration (auto-generated)

## Skills

The `skills/agent-deck/` directory contains a Claude Code skill that teaches Claude how to use the agent-deck CLI for managing sub-agents programmatically.

<!-- bv-agent-instructions-v1 -->

---

## Beads Workflow Integration

This project uses [beads_viewer](https://github.com/Dicklesworthstone/beads_viewer) for issue tracking. Issues are stored in `.beads/` and tracked in git.

### Essential Commands

```bash
# View issues (launches TUI - avoid in automated sessions)
bv

# CLI commands for agents (use these instead)
bd ready              # Show issues ready to work (no blockers)
bd list --status=open # All open issues
bd show <id>          # Full issue details with dependencies
bd create --title="..." --type=task --priority=2
bd update <id> --status=in_progress
bd close <id> --reason="Completed"
bd close <id1> <id2>  # Close multiple issues at once
bd sync               # Commit and push changes
```

### Workflow Pattern

1. **Start**: Run `bd ready` to find actionable work
2. **Claim**: Use `bd update <id> --status=in_progress`
3. **Work**: Implement the task
4. **Complete**: Use `bd close <id>`
5. **Sync**: Always run `bd sync` at session end

### Key Concepts

- **Dependencies**: Issues can block other issues. `bd ready` shows only unblocked work.
- **Priority**: P0=critical, P1=high, P2=medium, P3=low, P4=backlog (use numbers, not words)
- **Types**: task, bug, feature, epic, question, docs
- **Blocking**: `bd dep add <issue> <depends-on>` to add dependencies

### Session Protocol

**Before ending any session, run this checklist:**

```bash
git status              # Check what changed
git add <files>         # Stage code changes
bd sync                 # Commit beads changes
git commit -m "..."     # Commit code
bd sync                 # Commit any new beads changes
git push                # Push to remote
```

### Best Practices

- Check `bd ready` at session start to find available work
- Update status as you work (in_progress → closed)
- Create new issues with `bd create` when you discover tasks
- Use descriptive titles and set appropriate priority/type
- Always `bd sync` before ending session

<!-- end-bv-agent-instructions -->
