# GitPool

GitPool is a CLI + daemon tool for managing a pool of pre-initialized Git worktrees. It enables fast, disposable checkouts for builds, tests, and CI pipelines without repeated Git fetches.

## What is GitPool?

GitPool maintains a pool of pre-initialized Git worktrees that can be instantly claimed and released. Instead of waiting for `git clone` or `git fetch` operations, developers and CI systems get immediate access to ready-to-use worktrees.

## Why Use GitPool?

- **Instant checkouts**: No waiting for git operations - worktrees are pre-fetched and ready
- **Perfect for CI/CD**: Dramatically speed up build and test pipelines 
- **Resource efficient**: Worktrees share Git objects with the source repository
- **Safe isolation**: Each claimed worktree is independent and protected from updates
- **Automatic maintenance**: Background daemon keeps worktrees fresh and pool at capacity

## Installation

```bash
go install github.com/albertywu/gitpool/cmd@latest
```

## Quick Start

1. Start the daemon:
```bash
gitpool daemon start
```

2. Add a repository:
```bash
gitpool repo add my-app ~/repos/my-app --max 8 --default-branch develop
```

3. Claim a worktree:
```bash
gitpool claim --repo my-app
# Output (two lines):
# my-app-a91b6fc1
# /home/user/.gitpool/worktrees/my-app-a91b6fc1
```

4. Release a worktree:
```bash
gitpool release my-app-a91b6fc1-b837-4f76-93ef-37e4f5e37b31
```

5. Check pool status:
```bash
gitpool pool status
```

## Commands

### Daemon Management

- `gitpool daemon start` - Start the background daemon
- `gitpool daemon status` - Check daemon status

### Repository Management

- `gitpool repo add <name> <path>` - Register a Git repository
- `gitpool repo list` - List all registered repositories
- `gitpool repo remove <name>` - Remove a repository

### Worktree Operations

- `gitpool claim --repo <name>` - Claim an available worktree
- `gitpool release <worktree-id>` - Return a worktree to the pool
- `gitpool pool status` - Show pool usage statistics

## Examples

### CI/CD Pipeline Integration
```bash
# In your CI script
OUTPUT=$(gitpool claim --repo my-app)
WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)

cd "$WORKTREE_PATH"
make test
gitpool release $WORKTREE_ID
```

### Development Workflow
```bash
# Quick experimentation without affecting main workspace
OUTPUT=$(gitpool claim --repo my-project)
WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)

cd "$WORKTREE_PATH"
# ... make changes, test ideas ...
gitpool release $WORKTREE_ID
```

### Parallel Testing
```bash
# Run tests in parallel across multiple worktrees
for i in {1..4}; do
  OUTPUT=$(gitpool claim --repo my-app)
  WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)
  (cd "$WORKTREE_PATH" && make test-suite-$i) &
done
wait
```

## Configuration

Configuration is optional. gitpool uses `~/.gitpool/config.yaml` when present:

```yaml
# How often the daemon checks repositories (default: 1m)
reconciliation_interval: 1m

# Per-repository fetch intervals
repos:
  my-app:
    fetch_interval: 5m    # Check for updates every 5 minutes (default: 1h)
  legacy-service:
    fetch_interval: 30m   # Less frequent updates for stable repos (default: 1h)
```

## Documentation

- [Architecture](docs/architecture.md) - System design and components
- [Data Flow](docs/data-flow.md) - Reconciliation and update processes  
- [Storage](docs/storage.md) - File layout and database schema