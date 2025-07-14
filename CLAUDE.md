# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GitPool is a CLI + daemon tool for managing a pool of pre-initialized Git worktrees. It enables fast, disposable checkouts for builds, tests, and CI pipelines without repeated Git fetches.

## Build Commands

```bash
make build              # Build the gitpool binary
make test              # Run all tests (unit + integration)
make test-unit         # Run unit tests only
make test-integration  # Run integration tests
make fmt               # Format Go code with goimports
make lint              # Run golangci-lint
make install           # Install to /usr/local/bin
```

## Architecture

**Client-Server Model**:
- CLI client (`cmd/`) communicates with daemon via Unix socket
- Daemon (`daemon/`) manages worktree pool in background
- SQLite database stores persistent state (`db/store.go`)

**Key Components**:
- `cmd/commands/`: Individual CLI command implementations
- `daemon/reconciler.go`: Maintains pool health and updates
- `pool/`: Worktree allocation and lifecycle management
- `repo/`: Git repository operations
- `models/`: Core data structures (Repo, Worktree, Status)
- `ipc/`: Unix socket communication protocol

**Data Locations**:
- Database: `~/.gitpool/worktrees/gitpool.db`
- Socket: `~/.gitpool/worktrees/daemon.sock`
- Worktrees: `~/.gitpool/worktrees/<repo>/<id>`

## Testing

Integration tests in `tests/` create isolated environments. Run a single test:
```bash
go test ./tests -run TestSpecificFunction -v
```

Tests use temporary directories (`/tmp/gitpool-test-*`) and clean up automatically.

## Key Commands

- `gitpool start`: Start daemon
- `gitpool track <repo-name> <repo-path>`: Track repository (use --max flag for worktree count, defaults to 8)
- `gitpool untrack <repo-name>`: Stop tracking repository and clean up worktrees
- `gitpool refresh <repo-name>`: Manually fetch updates and refresh idle worktrees
- `gitpool use <repo> --branch <branch>`: Get worktree (returns JSON)
- `gitpool release <worktree-id>`: Return worktree to pool
- `gitpool show <worktree-id>`: Get details about a specific worktree (supports --format path)
- `gitpool list`: Show all worktrees

## Development Notes

1. **Branch uniqueness**: Each branch name must be unique per repository
2. **JSON output**: The `use` command returns JSON with worktree ID and path for scripting
3. **Error handling**: Use descriptive errors and proper cleanup in defer blocks
4. **Testing**: Always run `make test` before committing changes