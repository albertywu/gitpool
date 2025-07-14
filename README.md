# GitPool

GitPool is a CLI + daemon tool for managing a pool of pre-initialized Git worktrees. It enables fast, disposable checkouts for builds, tests, and CI pipelines without repeated Git fetches.

## What is GitPool?

GitPool maintains a pool of pre-initialized Git worktrees that can be instantly used and released. Instead of waiting on slow `git clone` or `git fetch` or `git checkout` operations, developers and CI systems get immediate access to ready-to-use worktrees.

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

3. Claim a worktree:
```bash
gitpool claim my-app --branch feature-xyz
# Output (JSON):
# {
#   "worktree_id": "a91b6fc1-1234-5678-90ab-cdef12345678",
#   "path": "/home/user/.gitpool/worktrees/my-app/a91b6fc1-1234-5678-90ab-cdef12345678"
# }

# To cd into the worktree:
cd $(gitpool claim my-app --branch feature-xyz | jq -r .path)

# Or save the worktree ID and query it later:
WORKTREE_ID=$(gitpool claim my-app --branch feature-xyz | jq -r .worktree_id)
cd $(gitpool show $WORKTREE_ID --format path)
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
- `gitpool refresh <name>` - Fetch updates and refresh idle worktrees
- `gitpool list` - List all worktrees with detailed status

- `gitpool claim <name> --branch <branch>` - Claim an available worktree with a unique branch name
- `gitpool release <worktree-id>` - Return a worktree to the pool
- `gitpool show <worktree-id>` - Get details about a specific worktree

## Examples

### CI/CD Pipeline Integration
```bash
# In your CI script
OUTPUT=$(gitpool claim my-app --branch "ci-run-${BUILD_ID}")
WORKTREE_ID=$(echo "$OUTPUT" | jq -r .worktree_id)
WORKTREE_PATH=$(echo "$OUTPUT" | jq -r .path)

cd "$WORKTREE_PATH"
make test
gitpool release $WORKTREE_ID
```

### Development Workflow
```bash
# Quick experimentation without affecting main workspace
OUTPUT=$(gitpool claim my-project --branch experiment-feature)
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
  WORKTREE_PATH=$(gitpool claim my-app --branch "test-suite-$i" | jq -r .path)
  (cd "$WORKTREE_PATH" && make test-suite-$i) &
done
wait
```

## Branch Management

GitPool requires a unique branch name when claiming a worktree:

- **Branch names must be unique** within a repository - no two active worktrees can use the same branch
- **Branch validation** ensures names follow Git conventions (no spaces, special characters, etc.)
- **Automatic cleanup** - branch associations are cleared when worktrees are released

The `gitpool list` command shows:
- **Sorting**: Claimed worktrees appear first, followed by unclaimed ones
- **Claimed worktrees**: Display the branch name in yellow as a clickable link
- **Unclaimed worktrees**: Display "UNCLAIMED" in gray as a clickable link
- **Claimed_at column**: Shows when a worktree was claimed, or "-" for unclaimed worktrees
- All links open the worktree directory when clicked (Cmd/Ctrl+click in supported terminals)

Example output:
```
ID                                    WORKTREE        REPO            CLAIMED_AT
────────────────────────────────────  ─────────────   ─────────────   ──────────────
a91b6fc1-1234-5678-90ab-cdef12345678  feature-xyz     backend-api     5m ago
c73d8ef3-3456-789a-12cd-ef3456789012  hotfix-123      backend-api     1h ago
e55fa0b5-5678-9abc-34ef-0b5678901234  experiment-ui   frontend-app    30m ago
b82c7de2-2345-6789-01bc-def234567890  UNCLAIMED       backend-api     -
d64e9fa4-4567-89ab-23de-fa4567890123  UNCLAIMED       frontend-app    -
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