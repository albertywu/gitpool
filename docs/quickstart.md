---
layout: default
title: Quick Start Guide
description: Get up and running with GitPool in minutes
---

# Quick Start Guide

Get GitPool running in your environment in just a few minutes.

## Installation

### Go Install (Recommended)
```bash
go install github.com/albertywu/gitpool/cmd@latest
```

### From Source
```bash
git clone https://github.com/albertywu/gitpool.git
cd gitpool
make build
sudo mv gitpool /usr/local/bin/
```

### Verify Installation
```bash
gitpool --help
```

## Basic Setup

### 1. Start the GitPool Daemon
```bash
gitpool start
```

The daemon runs in the background and manages your worktree pools. You can also run it in the foreground for debugging:

```bash
gitpool start --foreground
```

### 2. Add Your First Repository
```bash
gitpool track my-project ~/code/my-project --max 5 --default-branch main
```

**Parameters:**
- `my-project`: Pool name (used for claiming worktrees)
- `~/code/my-project`: Path to your Git repository
- `--max 5`: Maximum number of worktrees in the pool
- `--default-branch main`: Branch to initialize worktrees with

### 3. List Worktrees
```bash
gitpool list
```

You should see your repository's worktrees being initialized. GitPool will create worktrees in the background.

### 4. Claim Your First Worktree
```bash
gitpool claim my-project --branch feature-login
```

This returns two lines:
1. **Worktree ID**: Unique identifier for tracking
2. **Worktree Path**: Directory where you can work

Example output:
```
a91b6fc1-1234-5678-90ab-cdef12345678
/home/user/.gitpool/worktrees/my-project/a91b6fc1-1234-5678-90ab-cdef12345678
```

### 5. Work in Your Worktree
```bash
# Save the output for easy access
OUTPUT=$(gitpool claim my-project --branch feature-login)
WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)

# Change to the worktree directory
cd "$WORKTREE_PATH"

# Your worktree is ready! Make changes, run tests, etc.
git status
# You're on branch 'feature-login'
```

### 6. Release When Done
```bash
gitpool release $WORKTREE_ID
```

The worktree is cleaned and returned to the pool for reuse.

## Configuration (Optional)

Create `~/.gitpool/config.yaml` to customize behavior:

```yaml
# How often to check for updates (default: 1m)
reconciliation_interval: 30s

# Per-repository settings
repos:
  my-project:
    fetch_interval: 5m    # Fetch updates every 5 minutes
  legacy-app:
    fetch_interval: 1h    # Less frequent updates for stable repos
```

## Useful Commands

### View All Worktrees
```bash
gitpool list
```

Shows detailed status with clickable links to worktree directories.

### Stop the Daemon
```bash
gitpool stop
```

Or use `Ctrl+C` if running in foreground mode.

### Remove a Repository
```bash
gitpool untrack my-project
```

## Integration Examples

### Shell Script Integration
```bash
#!/bin/bash
set -e

# Claim a worktree
OUTPUT=$(gitpool claim my-project --branch "ci-build-$(date +%s)")
WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)

# Ensure cleanup on exit
trap "gitpool release $WORKTREE_ID" EXIT

# Work in the worktree
cd "$WORKTREE_PATH"
make test
make build

# Release is handled by trap
```

### CI/CD Pipeline (GitHub Actions)
```yaml
- name: Setup GitPool
  run: |
    gitpool start --foreground &
    gitpool track my-app . --max 3
    
- name: Run Tests
  run: |
    OUTPUT=$(gitpool claim my-app --branch "ci-${{ github.run_id }}")
    WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)
    cd "$WORKTREE_PATH"
    npm test
    gitpool release $(echo "$OUTPUT" | head -n1)
```

### Makefile Integration
```makefile
SHELL := /bin/bash

test-parallel:
	@for i in {1..4}; do \
		OUTPUT=$$(gitpool claim my-app --branch "test-$$i"); \
		WORKTREE_PATH=$$(echo "$$OUTPUT" | tail -n1); \
		(cd "$$WORKTREE_PATH" && make test-suite-$$i) & \
	done; \
	wait

.PHONY: test-parallel
```

## Troubleshooting

### Daemon Won't Start
- Check if another instance is running: `ps aux | grep gitpool`
- Verify permissions on `~/.gitpool/` directory
- Run with `--foreground` to see error messages

### Worktree Claim Fails
- Ensure the pool has available worktrees: `gitpool list`
- Check that branch name is unique and valid
- Verify repository is properly added: `gitpool list`

### Performance Issues
- Reduce `reconciliation_interval` in config
- Increase `--max` parameter when adding repositories
- Monitor disk space in `~/.gitpool/worktrees/`

## Next Steps

- [Read the full documentation](/docs)
- [Explore advanced examples](/examples)
- [Learn about GitPool's architecture](/docs/architecture)
- [Contribute on GitHub](https://github.com/albertywu/gitpool)

<style>
/* Quick Start Specific Styles */
.language-bash {
  background: #1a1a1a;
  color: #f8f8f2;
  padding: 1rem;
  border-radius: 8px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.9rem;
  overflow-x: auto;
  border-left: 4px solid #667eea;
}

.language-yaml {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-left: 4px solid #28a745;
}

h3 {
  color: #667eea;
  border-bottom: 2px solid #e9ecef;
  padding-bottom: 0.5rem;
  margin-top: 2rem;
}

strong {
  color: #495057;
}

code {
  background: #f8f9fa;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  font-size: 0.9em;
  color: #e83e8c;
}

blockquote {
  border-left: 4px solid #ffc107;
  background: #fff9c4;
  padding: 1rem;
  margin: 1rem 0;
}
</style>