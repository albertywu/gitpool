# GitPool

GitPool is a CLI + daemon tool for managing a pool of pre-initialized Git worktrees. It enables fast, disposable checkouts for builds, tests, and CI pipelines without repeated Git fetches.

## What is GitPool?

GitPool maintains a pool of pre-initialized Git worktrees that can be instantly used and released. Instead of waiting for `git clone` or `git fetch` operations, developers and CI systems get immediate access to ready-to-use worktrees.

## Why Use GitPool?

- **Instant checkouts**: No waiting for git operations - worktrees are pre-fetched and ready
- **Perfect for CI/CD**: Dramatically speed up build and test pipelines 
- **Resource efficient**: Worktrees share Git objects with the source repository
- **Safe isolation**: Each used worktree is independent and protected from updates
- **Automatic maintenance**: Background daemon keeps worktrees fresh and pool at capacity

## Installation

```bash
go install github.com/albertywu/gitpool/cmd@latest
```

## Quick Start

1. Start the daemon:
```bash
gitpool start
```

2. Add a repository:
```bash
gitpool track my-app ~/repos/my-app --max 8 --default-branch develop
```

3. Use a worktree:
```bash
gitpool use my-app --branch feature-xyz
# Output (JSON):
# {
#   "worktree_id": "a91b6fc1-1234-5678-90ab-cdef12345678",
#   "path": "/home/user/.gitpool/worktrees/my-app/a91b6fc1-1234-5678-90ab-cdef12345678"
# }

# To cd into the worktree:
cd $(gitpool use my-app --branch feature-xyz | jq -r .path)
```

4. Release a worktree when done:
```bash
gitpool release a91b6fc1-1234-5678-90ab-cdef12345678
```


## Commands

- `gitpool start` - Start the background daemon
- `gitpool stop` - Stop the daemon (or use Ctrl+C)

- `gitpool track <name> <path>` - Track a Git repository
- `gitpool untrack <name>` - Stop tracking a repository
- `gitpool list` - List all worktrees with detailed status

- `gitpool use <name> --branch <branch>` - Use an available worktree with a unique branch name
- `gitpool release <worktree-id>` - Return a worktree to the pool

## Examples

### CI/CD Pipeline Integration
```bash
# In your CI script
OUTPUT=$(gitpool use my-app --branch "ci-run-${BUILD_ID}")
WORKTREE_ID=$(echo "$OUTPUT" | jq -r .worktree_id)
WORKTREE_PATH=$(echo "$OUTPUT" | jq -r .path)

cd "$WORKTREE_PATH"
make test
gitpool release $WORKTREE_ID
```

### Development Workflow
```bash
# Quick experimentation without affecting main workspace
OUTPUT=$(gitpool use my-project --branch experiment-feature)
WORKTREE_ID=$(echo "$OUTPUT" | jq -r .worktree_id)
WORKTREE_PATH=$(echo "$OUTPUT" | jq -r .path)

cd "$WORKTREE_PATH"
# ... make changes, test ideas ...
gitpool release $WORKTREE_ID
```

### Parallel Testing
```bash
# Run tests in parallel across multiple worktrees
for i in {1..4}; do
  WORKTREE_PATH=$(gitpool use my-app --branch "test-suite-$i" | jq -r .path)
  (cd "$WORKTREE_PATH" && make test-suite-$i) &
done
wait
```

## Branch Management

GitPool requires a unique branch name when using a workspace:

- **Branch names must be unique** within a repository - no two active workspaces can use the same branch
- **Branch validation** ensures names follow Git conventions (no spaces, special characters, etc.)
- **Automatic cleanup** - branch associations are cleared when workspaces are released

The `gitpool list` command shows:
- **In-use workspaces**: Display the branch name in yellow as a clickable link
- **Unclaimed workspaces**: Display "UNCLAIMED" in gray as a clickable link
- All links open the workspace directory when clicked (Cmd/Ctrl+click in supported terminals)

Example output:
```
ID                                    WORKSPACE       REPO     STATUS    MAX  BRANCH    FETCH
────────────────────────────────────  ─────────────   ──────   ────────  ───  ────────  ──────
a91b6fc1-1234-5678-90ab-cdef12345678  feature-xyz     my-app   IN-USE    8    develop   1h
b82c7de2-2345-6789-01bc-def234567890  UNCLAIMED       my-app   IDLE      8    develop   1h
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