# GitPool

GitPool manages a pool of pre-initialized Git worktrees for instant checkouts without waiting for git operations.

## Installation

```bash
go install github.com/albertywu/gitpool/gp@latest
```

## Commands

```bash
gp start                              # Start the background daemon
gp stop                               # Stop the daemon
gp track <repo> <path>                # Track a Git repository
gp untrack <repo>                     # Stop tracking a repository
gp list                               # List all worktrees
gp claim <repo> <branch>              # Claim a worktree
gp release <worktree-id>              # Release a worktree back to the pool
gp show <worktree-id>                 # Show worktree details
gp refresh <repo>                     # Fetch updates and refresh idle worktrees
```

## Quick Start

1. **Start daemon**: `gp start`
2. **Track repository**: `gp track my-app ~/repos/my-app`
3. **Claim worktree**: `gp claim my-app feature-xyz`
4. **Work in worktree**: `cd $(gp claim my-app feature-xyz | jq -r .path)`
5. **Release when done**: `gp release <worktree-id>`

## Features

- **Instant checkouts** - Worktrees are pre-fetched and ready
- **Resource efficient** - Shared Git objects across worktrees
- **Automatic maintenance** - Background daemon keeps pool healthy
- **Branch isolation** - Unique branch names prevent conflicts
- **JSON output** - Easy integration with scripts and CI/CD

## Example: CI/CD Integration

```bash
# In your CI script
OUTPUT=$(gp claim my-app "ci-run-${BUILD_ID}")
WORKTREE_PATH=$(echo "$OUTPUT" | jq -r .path)
cd "$WORKTREE_PATH" && make test && cd -
gp release $(echo "$OUTPUT" | jq -r .worktree_id)
```