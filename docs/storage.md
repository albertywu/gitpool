# Storage

## File System Layout

All gitpool data is stored under `~/.gitpool/`:

- **Worktrees**: `~/.gitpool/worktrees/` (not configurable)
  - Each worktree is stored as: `~/.gitpool/worktrees/<repo-name>/<uuid>/`
  - Example: `~/.gitpool/worktrees/my-app/a91b6fc1-b837-4f76-93ef-37e4f5e37b31/`

- **Database**: `~/.gitpool/worktrees/gitpool.db`
  - SQLite database containing all metadata
  - Persists across daemon restarts

- **Socket**: `~/.gitpool/worktrees/daemon.sock`
  - Unix socket for IPC communication
  - Created when daemon starts, removed when it stops

- **Configuration**: `~/.gitpool/config.yaml`
  - Optional configuration file
  - Controls reconciliation and fetch intervals
  - If not present, defaults are used

## Database Schema

The SQLite database tracks:

### Repositories Table
- Repository name
- Source path
- Default branch
- Maximum worktrees
- Last fetch timestamp

### Worktrees Table
- Worktree ID (UUID)
- Repository name
- Worktree path
- Status (idle/in-use/corrupt)
- Branch name (when claimed)
- Created timestamp
- Last used timestamp

### Metadata
- Schema version
- Migration history

## Storage Requirements

- Each worktree is a full Git worktree (shares objects with source repo)
- Typical worktree size: metadata only (~100KB-1MB)
- Database size: minimal, grows with number of repositories and worktrees
- No cleanup needed - treefarm manages its own storage

## Persistence

- **Database persists**: Repository configurations, worktree states, last fetch times
- **Worktrees persist**: Physical worktree directories remain between daemon restarts
- **Configuration persists**: YAML config file is never modified by gitpool

## Backup and Recovery

To backup gitpool state:
1. Stop the daemon: `gitpool daemon stop`
2. Copy `~/.gitpool/worktrees/gitpool.db`
3. Optionally copy `~/.gitpool/config.yaml` (if it exists)

To restore:
1. Ensure daemon is stopped
2. Restore the database file
3. Start daemon: `gitpool daemon start`
4. Daemon will reconcile any missing worktrees automatically