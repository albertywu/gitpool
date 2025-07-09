# Treefarm

Treefarm is a CLI + daemon tool for managing a pool of pre-initialized Git worktrees. It enables fast, disposable checkouts for builds, tests, and CI pipelines without repeated Git fetches.

## What is Treefarm?

Treefarm maintains a pool of pre-initialized Git worktrees that can be instantly claimed and released. Instead of waiting for `git clone` or `git fetch` operations, developers and CI systems get immediate access to ready-to-use worktrees.

## Why Use Treefarm?

- **Instant checkouts**: No waiting for git operations - worktrees are pre-fetched and ready
- **Perfect for CI/CD**: Dramatically speed up build and test pipelines 
- **Resource efficient**: Worktrees share Git objects with the source repository
- **Safe isolation**: Each claimed worktree is independent and protected from updates
- **Automatic maintenance**: Background daemon keeps worktrees fresh and pool at capacity

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

## Examples

### CI/CD Pipeline Integration
```bash
# In your CI script
WORKTREE=$(treefarm claim --repo my-app --output-path)
cd "$WORKTREE"
make test
treefarm release $(basename "$WORKTREE")
```

### Development Workflow
```bash
# Quick experimentation without affecting main workspace
WORKTREE_ID=$(treefarm claim --repo my-project)
cd ~/.treefarm/worktrees/$WORKTREE_ID
# ... make changes, test ideas ...
treefarm release $WORKTREE_ID
```

### Parallel Testing
```bash
# Run tests in parallel across multiple worktrees
for i in {1..4}; do
  WORKTREE=$(treefarm claim --repo my-app --output-path)
  (cd "$WORKTREE" && make test-suite-$i) &
done
wait
```

## Configuration

Configure treefarm via `~/.treefarm/treefarm.yaml`:

```yaml
# How often the daemon checks repositories (default: 1m)
reconciliation_interval: 1m

# Per-repository fetch intervals
repos:
  my-app:
    fetch_interval: 5m    # Check for updates every 5 minutes
  legacy-service:
    fetch_interval: 30m   # Less frequent updates for stable repos
```

## Documentation

- [Architecture](docs/architecture.md) - System design and components
- [Data Flow](docs/data-flow.md) - Reconciliation and update processes  
- [Storage](docs/storage.md) - File layout and database schema