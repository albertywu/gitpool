# Treefarm

Treefarm is a CLI + daemon tool for managing a pool of pre-initialized Git worktrees. It enables fast, disposable checkouts for builds, tests, and CI pipelines without repeated Git fetches. Developers can instantly "claim" worktrees and "release" them back for reuse.

## Features

- Pre-initialized worktree pool management
- Fast worktree allocation without git fetch overhead
- Automatic cleanup and reset of released worktrees
- Background daemon with automatic reconciliation
- Support for multiple repositories with different configurations
- SQLite-based metadata persistence
- Worktrees stored in `~/.treefarm/worktrees`

## Installation

```bash
go install github.com/uber/treefarm/cmd@latest
```

## Quick Start

1. Start the daemon:
```bash
treefarm daemon start
```

2. Add a repository:
```bash
treefarm repo add my-app ~/repos/my-app --max 8 --default-branch develop
```

3. Claim a worktree:
```bash
# Get worktree name
treefarm claim --repo my-app

# Get full path
treefarm claim --repo my-app --output-path
```

4. Release a worktree:
```bash
treefarm release my-app-a91b6fc1-b837-4f76-93ef-37e4f5e37b31
```

5. Check pool status:
```bash
treefarm pool status
```

## Commands

### Daemon Management

- `treefarm daemon start` - Start the background daemon
- `treefarm daemon status` - Check daemon status

### Repository Management

- `treefarm repo add <name> <path>` - Register a Git repository
- `treefarm repo list` - List all registered repositories
- `treefarm repo remove <name>` - Remove a repository

### Worktree Operations

- `treefarm claim --repo <name>` - Claim an available worktree
- `treefarm release <worktree-id>` - Return a worktree to the pool
- `treefarm pool status` - Show pool usage statistics

## Configuration

Treefarm looks for configuration in `~/.treefarm/treefarm.yaml`. 

Configure the global reconciliation interval (how often the daemon checks for repository updates):

```yaml
reconciliation_interval: 1m  # Default: 1m (optional)
```

Configure per-repository fetch intervals:

```yaml
repos:
  git-spice:
    fetch_interval: 15m
  my-app:
    fetch_interval: 5m
```

When unset, repository fetch intervals default to 1 hour.

## Reconciliation

The reconciler runs continuously in the background to keep worktrees updated and maintain pool capacity. It operates on two interval levels:

### Global Reconciliation Interval (1m default)
The reconciler wakes up every minute (configurable via `reconciliation_interval` in config) and checks all registered repositories.

### Per-Repository Fetch Interval (1h default)
Each repository has its own fetch interval (configurable via `repos.<name>.fetch_interval` in config). During each reconciler run, only repositories that haven't been updated recently are processed.

**Last fetch times are persisted in the database**, so daemon restarts won't trigger unnecessary fetches - the system remembers when each repository was last updated.

### What Happens During Reconciliation

For each repository that's due for an update:

1. **Fetch main repository**: Runs `git fetch --all --prune` on the original repository to get latest changes
2. **Update idle worktrees**: Updates only **unclaimed** worktrees to the latest default branch
3. **Maintain capacity**: Creates new worktrees if under the configured maximum
4. **Clean up**: Removes corrupted worktrees and replaces them

### Safety

- **In-use worktrees are never touched** - Only idle (unclaimed) worktrees are updated
- **Your active work is protected** - Claimed worktrees remain exactly as you left them
- **Automatic cleanup** - Released worktrees are reset and cleaned before being returned to the pool

## Storage

- **Worktrees**: Stored in `~/.treefarm/worktrees/` (not configurable)
- **Database**: SQLite database at `~/.treefarm/worktrees/treefarm.db`
- **Socket**: IPC socket at `~/.treefarm/worktrees/daemon.sock`

## Architecture

- **Daemon**: Background service managing the worktree pool
- **Reconciler**: Ensures pool capacity and updates worktrees
- **IPC**: Unix socket communication between CLI and daemon
- **Storage**: SQLite database for metadata persistence in `~/.treefarm/worktrees`

### Worktree Allocation

When you add a repository, treefarm immediately creates all worktrees up to the configured maximum (`--max` flag). The reconciler continuously ensures the pool stays at capacity by creating new worktrees as needed and cleaning up corrupted ones.