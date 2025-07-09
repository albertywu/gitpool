# Storage

## File System Layout

All treefarm data is stored under `~/.treefarm/`:

- **Worktrees**: `~/.treefarm/worktrees/` (not configurable)
  - Each worktree is stored as: `~/.treefarm/worktrees/<repo-name>-<uuid>/`
  - Example: `~/.treefarm/worktrees/my-app-a91b6fc1-b837-4f76-93ef-37e4f5e37b31/`

- **Database**: `~/.treefarm/worktrees/treefarm.db`
  - SQLite database containing all metadata
  - Persists across daemon restarts

- **Socket**: `~/.treefarm/worktrees/daemon.sock`
  - Unix socket for IPC communication
  - Created when daemon starts, removed when it stops

- **Configuration**: `~/.treefarm/treefarm.yaml`
  - Optional configuration file
  - Controls reconciliation and fetch intervals

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
- Status (idle/in-use)
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
- **Configuration persists**: YAML config file is never modified by treefarm

## Backup and Recovery

To backup treefarm state:
1. Stop the daemon: `treefarm daemon stop`
2. Copy `~/.treefarm/worktrees/treefarm.db`
3. Optionally copy `~/.treefarm/treefarm.yaml`

To restore:
1. Ensure daemon is stopped
2. Restore the database file
3. Start daemon: `treefarm daemon start`
4. Daemon will reconcile any missing worktrees automatically