---
layout: default
title: Documentation
description: Complete documentation for GitPool including architecture, commands, and configuration
---

# Documentation

Comprehensive guide to GitPool's features, architecture, and usage patterns.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Commands Reference](#commands-reference)
- [Configuration](#configuration)
- [Architecture](#architecture)
- [Troubleshooting](#troubleshooting)
- [API Reference](#api-reference)

## Overview

GitPool is a CLI tool and daemon that manages pools of pre-initialized Git worktrees. It eliminates the wait time associated with `git clone`, `git fetch`, and `git checkout` operations by maintaining ready-to-use worktrees in the background.

### Key Concepts

- **Pool**: A collection of worktrees for a specific Git repository
- **Worktree**: An isolated Git working directory that shares objects with the source repository
- **Claim**: Taking a worktree from the pool for exclusive use
- **Release**: Returning a worktree to the pool for reuse
- **Daemon**: Background process that maintains pools and keeps worktrees fresh

## Installation

### Prerequisites
- Go 1.19 or later
- Git 2.5 or later
- Unix-like operating system (Linux, macOS)

### Install Methods

#### Go Install (Recommended)
```bash
go install github.com/albertywu/gitpool/cmd@latest
```

#### From Source
```bash
git clone https://github.com/albertywu/gitpool.git
cd gitpool
make build
sudo mv gitpool /usr/local/bin/
```

#### Homebrew (macOS)
```bash
# Coming soon
brew install gitpool
```

### Verify Installation
```bash
gitpool --version
gitpool --help
```

## Commands Reference

### Daemon Management

#### `gitpool start`
Start the GitPool daemon.

```bash
gitpool start [--foreground] [--config CONFIG_PATH]
```

**Options:**
- `--foreground`: Run in foreground instead of background
- `--config PATH`: Path to configuration file (default: `~/.gitpool/config.yaml`)

**Examples:**
```bash
# Start daemon in background
gitpool start

# Start in foreground for debugging
gitpool start --foreground

# Use custom config
gitpool start --config /etc/gitpool/config.yaml
```

#### `gitpool stop`
Stop the GitPool daemon.

```bash
gitpool stop
```

### Repository Management

#### `gitpool track`
Track a Git repository in GitPool.

```bash
gitpool track NAME PATH [OPTIONS]
```

**Arguments:**
- `NAME`: Pool name for the repository
- `PATH`: Path to the Git repository

**Options:**
- `--max INTEGER`: Maximum worktrees in pool (default: 5)
- `--default-branch BRANCH`: Default branch for worktrees (default: repository's HEAD)
- `--fetch-interval DURATION`: How often to fetch updates (default: 1h)

**Examples:**
```bash
# Basic usage
gitpool track my-app ~/code/my-app

# With custom settings
gitpool track web-service ~/repos/web-service --max 10 --default-branch develop

# Custom fetch interval
gitpool track api ~/work/api --fetch-interval 30m
```

#### `gitpool untrack`
Stop tracking a repository in GitPool.

```bash
gitpool untrack NAME [--force]
```

**Options:**
- `--force`: Remove even if worktrees are claimed

**Examples:**
```bash
# Stop tracking repository (fails if worktrees are claimed)
gitpool untrack my-app

# Force untracking
gitpool untrack my-app --force
```

#### `gitpool list`
List all worktrees with detailed information.

```bash
gitpool list [--format FORMAT] [--repo REPO]
```

**Options:**
- `--format`: Output format (`table`, `json`, `yaml`) (default: `table`)
- `--repo NAME`: Filter by repository name

**Output includes:**
- Worktree ID and status
- Associated branch name
- Claim time and duration
- Last fetch time
- Clickable directory links (in terminal)

### Worktree Operations

#### `gitpool claim`
Claim an available worktree.

```bash
gitpool claim REPO --branch BRANCH [OPTIONS]
```

**Arguments:**
- `REPO`: Repository pool name
- `--branch BRANCH`: Unique branch name for the worktree

**Options:**
- `--timeout DURATION`: Maximum time to wait for available worktree (default: 30s)

**Output:**
Returns two lines:
1. Worktree ID (for releasing)
2. Worktree path (for working)

**Examples:**
```bash
# Basic claim
gitpool claim my-app --branch feature-auth

# With timeout
gitpool claim my-app --branch hotfix-123 --timeout 60s

# Capture output
OUTPUT=$(gitpool claim my-app --branch test-run)
WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)
```

#### `gitpool release`
Release a claimed worktree back to the pool.

```bash
gitpool release WORKTREE_ID [--force]
```

**Arguments:**
- `WORKTREE_ID`: ID returned from `gitpool claim`

**Options:**
- `--force`: Release even if worktree has uncommitted changes

**Examples:**
```bash
# Normal release
gitpool release a91b6fc1-1234-5678-90ab-cdef12345678

# Force release (loses uncommitted changes)
gitpool release a91b6fc1-1234-5678-90ab-cdef12345678 --force
```

## Configuration

GitPool uses YAML configuration files. The default location is `~/.gitpool/config.yaml`.

### Configuration Structure

```yaml
# Global settings
reconciliation_interval: 1m  # How often daemon checks for work
log_level: info              # Log level (debug, info, warn, error)
data_dir: ~/.gitpool         # Data directory

# Per-repository settings
repos:
  my-app:
    fetch_interval: 5m       # Override default fetch interval
    max_worktrees: 8         # Override --max from add command
    default_branch: develop  # Override default branch
    
  legacy-service:
    fetch_interval: 2h       # Less frequent updates
    max_worktrees: 3         # Smaller pool
```

### Configuration Options

#### Global Settings

| Setting | Description | Default |
|---------|-------------|---------|
| `reconciliation_interval` | How often daemon checks repositories | `1m` |
| `log_level` | Logging verbosity | `info` |
| `data_dir` | Directory for GitPool data | `~/.gitpool` |
| `socket_path` | Unix socket for daemon communication | `~/.gitpool/socket` |

#### Repository Settings

| Setting | Description | Default |
|---------|-------------|---------|
| `fetch_interval` | How often to fetch updates | `1h` |
| `max_worktrees` | Maximum worktrees in pool | `5` |
| `default_branch` | Default branch for new worktrees | Repository HEAD |

### Environment Variables

GitPool respects these environment variables:

- `GITPOOL_CONFIG`: Path to configuration file
- `GITPOOL_DATA_DIR`: Data directory override
- `GITPOOL_LOG_LEVEL`: Log level override

## Architecture

### System Components

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   CLI Tool  │    │   Daemon    │    │  Worktrees  │
│             │    │             │    │             │
│ ┌─────────┐ │    │ ┌─────────┐ │    │ ┌─────────┐ │
│ │Commands │ │◄──►│ │ Pool    │ │◄──►│ │ Git     │ │
│ │         │ │    │ │Manager  │ │    │ │ Repos   │ │
│ └─────────┘ │    │ └─────────┘ │    │ └─────────┘ │
│             │    │             │    │             │
│ ┌─────────┐ │    │ ┌─────────┐ │    │ ┌─────────┐ │
│ │ Client  │ │    │ │Reconciler│ │    │ │ Work    │ │
│ │ API     │ │    │ │         │ │    │ │ Dirs    │ │
│ └─────────┘ │    │ └─────────┘ │    │ └─────────┘ │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       └─────── Unix Socket ─────────────────────┘
```

### Data Flow

1. **Initialization**: Daemon creates initial worktrees for each repository
2. **Background Fetch**: Reconciler periodically fetches updates from origin
3. **Claim Request**: Client requests worktree for specific branch
4. **Worktree Assignment**: Daemon assigns available worktree and checks out branch
5. **User Work**: User works in isolated worktree
6. **Release**: Worktree is cleaned and returned to pool

### Storage Layout

```
~/.gitpool/
├── config.yaml          # Configuration file
├── gitpool.db           # SQLite database
├── socket               # Unix socket for IPC
├── logs/                # Daemon logs
│   └── gitpool.log
└── worktrees/           # Worktree storage
    ├── repo-1/
    │   ├── uuid1/       # Individual worktree
    │   ├── uuid2/
    │   └── uuid3/
    └── repo-2/
        ├── uuid4/
        └── uuid5/
```

### Database Schema

GitPool uses SQLite for persistence:

```sql
-- Repository configurations
CREATE TABLE repos (
    name TEXT PRIMARY KEY,
    path TEXT NOT NULL,
    max_worktrees INTEGER NOT NULL,
    default_branch TEXT NOT NULL,
    fetch_interval INTEGER NOT NULL
);

-- Worktree instances
CREATE TABLE worktrees (
    id TEXT PRIMARY KEY,
    repo_name TEXT NOT NULL,
    path TEXT NOT NULL,
    branch TEXT,
    claimed_at DATETIME,
    created_at DATETIME NOT NULL,
    last_fetched DATETIME,
    FOREIGN KEY (repo_name) REFERENCES repos(name)
);
```

## Troubleshooting

### Common Issues

#### Daemon Won't Start

**Symptoms:**
- `gitpool start` fails
- "Address already in use" errors

**Solutions:**
```bash
# Check if daemon is already running
ps aux | grep gitpool

# Check socket file
ls -la ~/.gitpool/socket

# Remove stale socket
rm ~/.gitpool/socket

# Start with debug logging
gitpool start --foreground --log-level debug
```

#### Claim Operations Fail

**Symptoms:**
- "No worktrees available" errors
- Timeouts during claim

**Solutions:**
```bash
# Check worktree list
gitpool list

# Increase pool size
gitpool untrack my-app
gitpool track my-app ~/repos/my-app --max 10

# Check for stale claims
gitpool list

# Force release stale worktrees
gitpool release WORKTREE_ID --force
```

#### Performance Issues

**Symptoms:**
- Slow claim operations
- High disk usage
- Memory leaks

**Solutions:**
```bash
# Reduce reconciliation frequency
echo "reconciliation_interval: 5m" >> ~/.gitpool/config.yaml

# Clean up worktree storage
du -sh ~/.gitpool/worktrees/

# Monitor daemon logs
tail -f ~/.gitpool/logs/gitpool.log

# Restart daemon periodically
gitpool stop && gitpool start
```

#### Git Operation Failures

**Symptoms:**
- Fetch errors in logs
- Checkout failures
- Permission denied errors

**Solutions:**
```bash
# Check Git configuration
git config --global user.name
git config --global user.email

# Verify repository access
cd ~/repos/my-app && git fetch

# Fix permissions
chmod -R u+rw ~/.gitpool/worktrees/

# Update Git credentials
git credential-manager-core configure
```

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
# Start daemon with debug logging
gitpool start --foreground --log-level debug

# Check logs
tail -f ~/.gitpool/logs/gitpool.log

# Enable Git debug output
export GIT_CURL_VERBOSE=1
export GIT_TRACE=1
```

### Health Checks

Monitor GitPool health:

```bash
#!/bin/bash
# health-check.sh

# Check daemon status
if ! pgrep -f "gitpool.*start" > /dev/null; then
    echo "❌ Daemon not running"
    exit 1
fi

# Check socket communication
if ! gitpool list > /dev/null 2>&1; then
    echo "❌ Cannot communicate with daemon"
    exit 1
fi

# Check disk space
USAGE=$(df ~/.gitpool | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$USAGE" -gt 90 ]; then
    echo "⚠️ Disk usage high: ${USAGE}%"
fi

# Check pool availability
AVAILABLE=$(gitpool list --format json | jq '[.[] | select(.status == "idle")] | length')
if [ "$AVAILABLE" -eq 0 ]; then
    echo "⚠️ No available worktrees"
fi

echo "✅ GitPool healthy"
```

## API Reference

### REST API (Future)

GitPool will support REST API in future versions:

```bash
# Get pool status
curl http://localhost:8080/api/v1/pools

# Claim worktree
curl -X POST http://localhost:8080/api/v1/pools/my-app/claim \
     -d '{"branch": "feature-123"}'

# Release worktree
curl -X DELETE http://localhost:8080/api/v1/worktrees/WORKTREE_ID
```

### Library Usage (Go)

Use GitPool as a Go library:

```go
package main

import (
    "context"
    "github.com/albertywu/gitpool/pkg/client"
)

func main() {
    // Connect to daemon
    client, err := client.New("~/.gitpool/socket")
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    // Claim worktree
    worktree, err := client.Claim(context.Background(), "my-app", "feature-branch")
    if err != nil {
        panic(err)
    }
    
    // Use worktree
    fmt.Printf("Worktree: %s\n", worktree.Path)
    
    // Release when done
    err = client.Release(context.Background(), worktree.ID)
    if err != nil {
        panic(err)
    }
}
```

### Integration Scripts

Common integration patterns:

```bash
# Wrapper function for safe worktree operations
claim_and_run() {
    local repo="$1"
    local branch="$2"
    shift 2
    local command="$@"
    
    local output
    output=$(gitpool claim "$repo" --branch "$branch") || return 1
    
    local worktree_id=$(echo "$output" | head -n1)
    local worktree_path=$(echo "$output" | tail -n1)
    
    trap "gitpool release $worktree_id" EXIT
    
    cd "$worktree_path"
    eval "$command"
}

# Usage
claim_and_run my-app test-$(date +%s) "make test && make build"
```

---

For more information, visit our [GitHub repository](https://github.com/albertywu/gitpool) or check out the [examples page](/examples).

<style>
/* Documentation styling */
.table-responsive {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0;
  background: white;
}

th, td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid #e9ecef;
}

th {
  background: #f8f9fa;
  font-weight: 600;
  color: #495057;
}

tr:hover {
  background: #f8f9fa;
}

/* Section navigation */
h2 {
  color: #667eea;
  border-bottom: 3px solid #e9ecef;
  padding-bottom: 0.5rem;
  margin-top: 3rem;
}

h3 {
  color: #495057;
  border-left: 4px solid #667eea;
  padding-left: 1rem;
  margin-top: 2rem;
}

/* Code blocks */
.language-bash,
.language-yaml,
.language-sql,
.language-go {
  background: #1a1a1a;
  color: #f8f8f2;
  padding: 1rem;
  border-radius: 8px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.9rem;
  overflow-x: auto;
  margin: 1rem 0;
}

.language-yaml {
  background: #f8f9fa;
  color: #333;
  border: 1px solid #e9ecef;
}

/* Callouts */
blockquote {
  border-left: 4px solid #17a2b8;
  background: #d1ecf1;
  padding: 1rem;
  margin: 1rem 0;
  border-radius: 0 8px 8px 0;
}

/* Architecture diagrams */
pre.ascii {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  padding: 1rem;
  font-family: monospace;
  white-space: pre;
  overflow-x: auto;
}

/* Table of contents */
ul {
  list-style-type: none;
  padding-left: 1rem;
}

ul li a {
  color: #667eea;
  text-decoration: none;
  padding: 0.25rem 0;
  display: block;
}

ul li a:hover {
  color: #495057;
  text-decoration: underline;
}

/* Command options formatting */
dt {
  font-weight: 600;
  color: #495057;
  margin-top: 0.5rem;
}

dd {
  margin-left: 1rem;
  margin-bottom: 0.5rem;
  color: #666;
}
</style>