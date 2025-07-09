# Architecture

GitPool follows a client-server architecture with persistent state management:

## Core Components

### Daemon
- Background service managing the worktree pool
- Runs continuously to maintain pool health
- Handles all worktree lifecycle operations
- Communicates via Unix socket IPC

### Reconciler
- Ensures pool capacity meets configured maximums
- Updates idle worktrees with latest changes
- Cleans up corrupted or invalid worktrees
- Operates on configurable intervals

### IPC (Inter-Process Communication)
- Unix socket communication between CLI and daemon
- Located at `~/.gitpool/worktrees/daemon.sock`
- Enables fast, secure local communication
- Protocol: JSON-RPC style messages

### Storage Layer
- SQLite database for metadata persistence
- Tracks worktree state, claims, and repository configurations
- Located at `~/.gitpool/worktrees/gitpool.db`
- Ensures state survives daemon restarts

## Worktree Lifecycle

### Creation
When you add a repository, gitpool immediately creates all worktrees up to the configured maximum (`--max` flag). Each worktree is:
- Created as a Git worktree of the source repository
- Initialized with the default branch
- Registered in the database with "idle" status

### Allocation
When a client claims a worktree:
1. Daemon finds an idle worktree for the requested repository
2. Marks it as "in-use" in the database
3. Returns the worktree path to the client
4. Worktree remains untouched during use

### Release
When a client releases a worktree:
1. Daemon marks it as "idle" in the database
2. Worktree is cleaned and reset in the background
3. Becomes available for future claims

### Maintenance
The reconciler continuously:
- Creates new worktrees if under capacity
- Updates idle worktrees to latest commits
- Removes and replaces corrupted worktrees
- Never touches in-use worktrees