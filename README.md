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

## Features

- **Instant checkouts** - Worktrees are pre-fetched and ready
- **Resource efficient** - Shared Git objects across worktrees
- **Automatic maintenance** - Background daemon keeps pool healthy
- **Branch isolation** - Unique branch names prevent conflicts
- **JSON output** - Easy integration with scripts and CI/CD


## Example: Run Agent in Worktree

```bash
# In your Agent script
OUTPUT=$(gp claim my-app "fix-the-thing")
WORKTREE_PATH=$(echo "$OUTPUT" | jq -r .path)
cd "$WORKTREE_PATH" && claude -p "fix the thing" && cd -
gp release $(echo "$OUTPUT" | jq -r .worktree_id)
```

## Example: CI/CD Integration

```bash
# In your CI script
OUTPUT=$(gp claim my-app "ci-run-${BUILD_ID}")
WORKTREE_PATH=$(echo "$OUTPUT" | jq -r .path)
cd "$WORKTREE_PATH" && make test && cd -
gp release $(echo "$OUTPUT" | jq -r .worktree_id)
```
