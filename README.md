# Treefarm

Treefarm is a CLI + daemon tool for managing a pool of pre-initialized Git worktrees. It enables fast, disposable checkouts for builds, tests, and CI pipelines without repeated Git fetches. Developers can instantly "claim" worktrees and "release" them back for reuse.

## Features

- Pre-initialized worktree pool management
- Fast worktree allocation without git fetch overhead
- Automatic cleanup and reset of released worktrees
- Background daemon with automatic reconciliation
- Support for multiple repositories with different configurations
- SQLite-based metadata persistence

## Installation

```bash
go install github.com/uber/treefarm/cmd@latest
```

## Quick Start

1. Start the daemon:
```bash
treefarm daemon start --workdir /data/wtpool --fetch-interval 15m
```

2. Add a repository:
```bash
treefarm repo add my-app ~/repos/my-app --max 8 --default-branch develop --fetch-interval 5m
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

Treefarm looks for configuration in `~/.treefarm/treefarm.yaml`. Example:

```yaml
work_dir: /data/wtpool
fetch_interval: 15m
```

## Architecture

- **Daemon**: Background service managing the worktree pool
- **Reconciler**: Ensures pool capacity and updates worktrees
- **IPC**: Unix socket communication between CLI and daemon
- **Storage**: SQLite database for metadata persistence