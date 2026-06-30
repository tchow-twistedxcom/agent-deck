# Fork Maintenance Guide

This repo is a fork of [asheshgoplani/agent-deck](https://github.com/asheshgoplani/agent-deck)
carrying Twisted X customizations. This document is the single source of truth for how the
fork is maintained.

## Branch model

| Branch | Role | Rules |
|--------|------|-------|
| `main` | Pure upstream mirror + 2 fork-only workflow files | Never commit here. Force-pushed daily by `upstream-sync.yml`. Locally treat as disposable: `git fetch && git branch -f main origin/main`. |
| `custom/dev` | The product branch: upstream + our patch set | All fork work lands here. Tracks `origin/custom/dev`. |

## Upstream tracking strategy: MERGE, not rebase

As of the v1.9.54 integration (June 2026), `custom/dev` uses **merge-based tracking**:

```bash
git fetch upstream
git checkout custom/dev
git merge upstream/main
# Resolve conflicts once; rerere replays previous resolutions
git push origin custom/dev
```

Do NOT rebase `custom/dev` onto upstream. The branch history contains upstream merge
commits; rebasing across them recreates the "rebase artifact" cleanup commits that
polluted the history during the rebase era (pre-v1.9.54). Merge keeps conflict
resolution to once per sync, needs no force-push, and preserves true history.

`git rerere` is enabled in this clone so recurring conflicts (the tmux/C-q keybinding
area is the usual hotspot) only need resolving once.

**Cadence**: sync on every upstream release tag, or when the daily `upstream-sync.yml`
run opens/updates the "custom/dev needs upstream merge" issue. Smaller gaps mean
smaller conflicts.

## Automation

`/.github/workflows/upstream-sync.yml` (fork-only, self-healing) runs daily at 9am UTC:

1. Hard-resets `main` to `upstream/main`, re-adds the fork-only workflow files
   (`upstream-sync.yml`, `ci.yml`), force-pushes with lease.
2. Measures `custom/dev` divergence and opens/updates a labeled issue
   (`upstream-sync`) with merge instructions when behind.

The canonical copy of these workflow files lives on `main`; the sync job re-seeds them
from main's checkout each run. Changes to them must be pushed to BOTH `custom/dev` and
`main` or the next sync silently reverts them on main.

`/.github/workflows/ci.yml` (fork-only) builds and tests pushes/PRs to `main`,
`custom/dev`, and `feature/*`.

## Patch inventory (as of v1.10.7 + 2)

Grouped by feature. "Upstream?" = candidate for submitting upstream; every patch
upstreamed is permanent merge burden removed.

| Feature | Commits (representative) | Upstream? |
|---------|--------------------------|-----------|
| Happy wrapper for Claude/Codex sessions (incl. resume path) | `5b96d734`, `dfa3d5a6` | No, org-specific |
| Block happy+chrome combo at creation/edit; strip --chrome from command build | `5fdce325` | No, org-specific (happy wrapper is org-specific) |
| OAuth support for HTTP MCP servers | `efd4b6ae` | **Yes, strong candidate** |
| Worktree: branch reuse + fzf picker, generated-name placeholder, parent-session symlink for resume | `ed917967`, `44f25141`, `f30747ce` | **Yes, strong candidate** |
| Auto-generated session names as dialog placeholders | `ef84c4b3` | Yes |
| Ctrl+Q detach / tmux keybinding behavior (C-q gate on socket isolation, no root binding, MakeRaw handling, extended-keys per-session, terminal-features dedup) | `37dedc9a`, `6bb20d54`, `dff4ba1a`, `6bd1ec93`, `7f8c524e`, `9744b76e`, `ce713a2e`, `fb6d29c5` | Partially; behavior preference, propose upstream as config flag |
| Claude session ID validation (corruption guard) | `37a3f509` | Yes, check if upstream fixed independently |
| Cost dashboard on C key, $ restored for error filter | `d9766925` | No, keybind preference |
| `sort_by_actionable` off switch (shim over upstream groupSortMode) | `61e8857a` | Absorbed: upstream added groupSortMode string; our SetSortByActionable is now a shim |
| Clear stale `AGENTDECK_PROFILE` from tmux global env | `abb0f757` | Yes |
| Fork validation helper (`sessionFileFoundButEmpty`) | `4642d8f8` | No, fork test support |
| Termius emoji-wide row clamping (terminalDrawWidth) | `ba1b4243` | No, Termius-specific; upstream uses ansi.StringWidth |
| Repo hygiene (.beads gitignore, fork CI files) | `5aa3b339`, `74e70078` | No, fork-only |

Integration-noise commits (`95912e58`, `48cbeb04`, `a2a3e5b8`, `ca3b1bcd`) are
leftovers from the rebase era; merge-based tracking stops producing these.

Known new upstream test failures in this environment (not caused by fork patches):
- `TestCanRestartCursor`: requires `cursor` binary, not present on this host.
- `TestCtrlS_NewDialogOpen_DoesNotOpenSwitcher`: wiring for the new session-switcher
  feature (#1411) interacts with the test harness in a way not yet diagnosed.

When adding a new customization: add a row here in the same commit. When upstream
absorbs a feature, delete the row and note the upstream PR.

## Build discipline

Install only from a clean, committed tree so every running binary maps to a commit:

```bash
git status --short   # must be empty
make install-user
agent-deck --version # must NOT say -dirty
```

## Backups

Pre-sync snapshots are kept as tags (`fork-backup/<version>`), not branches.
Keep the last two; delete older ones. Merge-based tracking makes these mostly
redundant (no history rewriting), they are cheap insurance only.

Current tags: `fork-backup/v1.9.54` (pre-v1.10.7 merge), `fork-backup/pre-v1.9.54`.
Older tags (`fork-backup/before-rebase-2026-03`, `fork-backup/origin-pre-v1.9.54-push`,
`fork-backup/pre-v1.9.31`, `fork-backup/pre-v1.9.44`) can be deleted when convenient.
