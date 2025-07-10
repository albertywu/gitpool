# Data Flow and Reconciliation

## Reconciliation Process

The reconciler runs continuously in the background to keep worktrees updated and maintain pool capacity. It operates on two interval levels:

### Global Reconciliation Interval (1m default)
The reconciler wakes up every minute (configurable via `reconciliation_interval` in config) and checks all registered repositories.

### Per-Repository Fetch Interval (1h default)
Each repository has its own fetch interval (configurable via `repos.<n>.fetch_interval` in config). During each reconciler run, only repositories that haven't been updated recently are processed.

**Last fetch times are persisted in the database**, so daemon restarts won't trigger unnecessary fetches - the system remembers when each repository was last updated.

## Reconciliation Steps

For each repository that's due for an update:

1. **Fetch main repository**: Runs `git fetch --all --prune` on the original repository to get latest changes
2. **Update idle worktrees**: Updates only **unclaimed** worktrees to the latest default branch
3. **Maintain capacity**: Creates new worktrees if under the configured maximum
4. **Clean up**: Removes corrupted worktrees and replaces them

## Data Flow

### Claim Flow
```
CLI Client (with branch) → IPC Socket → Daemon → Validate Branch Name → Check Branch Uniqueness → 
Database Query → Find Idle Worktree → Mark as In-Use + Set Branch → Return Path
```

### Release Flow
```
CLI Client → IPC Socket → Daemon → Database Update → Mark as Idle + Clear Branch → Background Cleanup Task
```

### Reconciliation Flow
```
Timer Trigger → Check Repository → Fetch Updates → Update Idle Worktrees → Create Missing Worktrees → Update Database
```

## Safety Guarantees

- **In-use worktrees are never touched** - Only idle (unclaimed) worktrees are updated
- **Your active work is protected** - Claimed worktrees remain exactly as you left them
- **Automatic cleanup** - Released worktrees are reset and cleaned before being returned to the pool
- **Atomic operations** - Database transactions ensure consistent state
- **Graceful degradation** - If a worktree is corrupted, it's removed and replaced automatically

## Configuration

### Global Configuration
Optional file located at `~/.gitpool/config.yaml`:
```yaml
reconciliation_interval: 1m  # How often reconciler runs
```

### Per-Repository Configuration
```yaml
repos:
  git-spice:
    fetch_interval: 15m
  my-app:
    fetch_interval: 5m
```

When unset, repository fetch intervals default to 1 hour.

If no config file is present, all settings use their defaults:
- `reconciliation_interval`: 1 minute
- Repository fetch intervals: 1 hour