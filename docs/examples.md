---
layout: default
title: Examples & Use Cases
description: Real-world examples of GitPool in CI/CD pipelines, development workflows, and automation
---

# Examples & Use Cases

Discover how GitPool accelerates development workflows across different scenarios.

## CI/CD Pipeline Integration

### GitHub Actions
Speed up your GitHub Actions workflows by eliminating git clone overhead:

```yaml
name: Fast CI with GitPool
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Install GitPool
        run: go install github.com/albertywu/gitpool/cmd@latest
        
      - name: Setup GitPool
        run: |
          gitpool start --foreground &
          sleep 2
          gitpool track repo . --max 4 --default-branch main
          
      - name: Run Tests
        run: |
          OUTPUT=$(gitpool claim repo --branch "ci-${{ github.run_id }}")
          WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
          WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)
          
          cd "$WORKTREE_PATH"
          npm ci
          npm test
          
          gitpool release $WORKTREE_ID
          
      - name: Stop GitPool
        run: gitpool stop
```

### Jenkins Pipeline
```groovy
pipeline {
    agent any
    stages {
        stage('Setup') {
            steps {
                sh 'gitpool start --foreground &'
                sh 'gitpool track ${JOB_NAME} ${WORKSPACE} --max 5'
            }
        }
        stage('Test') {
            parallel {
                stage('Unit Tests') {
                    steps {
                        script {
                            def output = sh(
                                script: "gitpool claim ${JOB_NAME} --branch unit-${BUILD_ID}",
                                returnStdout: true
                            ).trim().split('\n')
                            def worktreeId = output[0]
                            def worktreePath = output[1]
                            
                            dir(worktreePath) {
                                sh 'make test-unit'
                            }
                            
                            sh "gitpool release ${worktreeId}"
                        }
                    }
                }
                stage('Integration Tests') {
                    steps {
                        script {
                            def output = sh(
                                script: "gitpool claim ${JOB_NAME} --branch integration-${BUILD_ID}",
                                returnStdout: true
                            ).trim().split('\n')
                            def worktreeId = output[0]
                            def worktreePath = output[1]
                            
                            dir(worktreePath) {
                                sh 'make test-integration'
                            }
                            
                            sh "gitpool release ${worktreeId}"
                        }
                    }
                }
            }
        }
    }
    post {
        always {
            sh 'gitpool stop'
        }
    }
}
```

### GitLab CI
```yaml
variables:
  GITPOOL_CONFIG: ~/.gitpool/config.yaml

before_script:
  - go install github.com/albertywu/gitpool/cmd@latest
  - gitpool start --foreground &
  - gitpool track project $CI_PROJECT_DIR --max 3

stages:
  - test
  - build
  - deploy

test:
  stage: test
  script:
    - OUTPUT=$(gitpool claim project --branch "test-$CI_PIPELINE_ID")
    - WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
    - WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)
    - cd "$WORKTREE_PATH"
    - make test
    - gitpool release $WORKTREE_ID
  parallel: 3

after_script:
  - gitpool stop
```

## Development Workflows

### Quick Feature Experimentation
```bash
#!/bin/bash
# quick-experiment.sh - Quickly test ideas without affecting main workspace

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <experiment-name>"
    exit 1
fi

EXPERIMENT_NAME="experiment-$1-$(date +%s)"

# Claim a worktree
OUTPUT=$(gitpool claim my-project --branch "$EXPERIMENT_NAME")
WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)

echo "🧪 Starting experiment: $EXPERIMENT_NAME"
echo "📁 Worktree: $WORKTREE_PATH"
echo "🆔 ID: $WORKTREE_ID"

# Setup cleanup on exit
cleanup() {
    echo "🧹 Cleaning up experiment..."
    gitpool release $WORKTREE_ID
    echo "✅ Experiment cleaned up"
}
trap cleanup EXIT

# Open in your editor
cd "$WORKTREE_PATH"
code . # or your preferred editor

# Keep shell open for experimentation
exec bash
```

### Code Review Workflow
```bash
#!/bin/bash
# review-pr.sh - Review pull requests in isolated environments

PR_BRANCH="$1"
REVIEW_ID="review-$(whoami)-$(date +%s)"

if [ -z "$PR_BRANCH" ]; then
    echo "Usage: $0 <pr-branch-name>"
    exit 1
fi

# Claim worktree
OUTPUT=$(gitpool claim main-repo --branch "$REVIEW_ID")
WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)

# Setup cleanup
trap "gitpool release $WORKTREE_ID" EXIT

cd "$WORKTREE_PATH"

# Fetch and checkout PR branch
git fetch origin "$PR_BRANCH:$PR_BRANCH"
git checkout "$PR_BRANCH"

echo "📝 Reviewing PR branch: $PR_BRANCH"
echo "📁 Review environment: $WORKTREE_PATH"

# Run your review workflow
make lint
make test
make build

# Open for manual review
code .
echo "Press Enter when review is complete..."
read
```

### Multi-Version Testing
```bash
#!/bin/bash
# test-versions.sh - Test against multiple dependency versions

VERSIONS=("node:16" "node:18" "node:20")
PROJECT_NAME="web-app"

for version in "${VERSIONS[@]}"; do
    echo "🧪 Testing with $version"
    
    OUTPUT=$(gitpool claim $PROJECT_NAME --branch "test-$version-$(date +%s)")
    WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
    WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)
    
    (
        cd "$WORKTREE_PATH"
        
        # Use Docker to test with specific version
        docker run --rm -v "$(pwd):/app" -w /app $version bash -c "
            npm ci
            npm test
        "
        
        gitpool release $WORKTREE_ID
    ) &
done

wait
echo "✅ All version tests completed"
```

## Build & Release Automation

### Multi-Platform Builds
```bash
#!/bin/bash
# build-platforms.sh - Build for multiple platforms in parallel

PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")
PROJECT="go-cli"

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    echo "🏗️ Building for $GOOS/$GOARCH"
    
    OUTPUT=$(gitpool claim $PROJECT --branch "build-$GOOS-$GOARCH-$(date +%s)")
    WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
    WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)
    
    (
        cd "$WORKTREE_PATH"
        
        export GOOS=$GOOS
        export GOARCH=$GOARCH
        
        go build -o "dist/${PROJECT}-${GOOS}-${GOARCH}" ./cmd
        
        gitpool release $WORKTREE_ID
    ) &
done

wait
echo "✅ All platform builds completed"
```

### Release Pipeline
```bash
#!/bin/bash
# release.sh - Automated release workflow

VERSION="$1"
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

PROJECT="my-app"
RELEASE_BRANCH="release-$VERSION-$(date +%s)"

# Claim dedicated release environment
OUTPUT=$(gitpool claim $PROJECT --branch "$RELEASE_BRANCH")
WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)

# Ensure cleanup
trap "gitpool release $WORKTREE_ID" EXIT

cd "$WORKTREE_PATH"

echo "🚀 Starting release $VERSION"

# Update version
sed -i "s/\"version\": \".*\"/\"version\": \"$VERSION\"/" package.json

# Run full test suite
echo "🧪 Running tests..."
npm test

# Build production assets
echo "🏗️ Building..."
npm run build

# Create release commit
git add .
git commit -m "Release $VERSION"
git tag "v$VERSION"

# Push to release branch
git push origin "$RELEASE_BRANCH"
git push origin "v$VERSION"

echo "✅ Release $VERSION completed"
echo "📦 Tagged as v$VERSION"
echo "🌿 Released from branch $RELEASE_BRANCH"
```

## Performance Testing

### Load Testing with Multiple Workers
```bash
#!/bin/bash
# load-test.sh - Run load tests from multiple isolated environments

WORKERS=5
PROJECT="api-server"

echo "🚀 Starting load test with $WORKERS workers"

for i in $(seq 1 $WORKERS); do
    OUTPUT=$(gitpool claim $PROJECT --branch "load-test-worker-$i-$(date +%s)")
    WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
    WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)
    
    echo "👷 Starting worker $i in $WORKTREE_PATH"
    
    (
        cd "$WORKTREE_PATH"
        
        # Start API server on unique port
        PORT=$((8000 + i))
        npm start -- --port $PORT &
        SERVER_PID=$!
        
        # Wait for server to start
        sleep 5
        
        # Run load test
        ab -n 1000 -c 10 "http://localhost:$PORT/api/health"
        
        # Cleanup
        kill $SERVER_PID
        gitpool release $WORKTREE_ID
    ) &
done

wait
echo "✅ Load testing completed"
```

### Database Migration Testing
```bash
#!/bin/bash
# test-migrations.sh - Test database migrations safely

PROJECT="backend"
DB_VERSIONS=("postgres:12" "postgres:13" "postgres:14" "postgres:15")

for db_version in "${DB_VERSIONS[@]}"; do
    echo "🗄️ Testing migrations with $db_version"
    
    OUTPUT=$(gitpool claim $PROJECT --branch "migration-test-$db_version-$(date +%s)")
    WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
    WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)
    
    (
        cd "$WORKTREE_PATH"
        
        # Start database
        docker run -d --name "test-db-$WORKTREE_ID" \
            -e POSTGRES_PASSWORD=test \
            -p "5432" \
            $db_version
        
        # Wait for DB to be ready
        sleep 10
        
        # Run migrations
        make migrate-up
        
        # Run tests
        make test-db
        
        # Cleanup
        docker stop "test-db-$WORKTREE_ID"
        docker rm "test-db-$WORKTREE_ID"
        
        gitpool release $WORKTREE_ID
    ) &
done

wait
echo "✅ Migration testing completed"
```

## Advanced Automation

### Auto-scaling Test Environment
```bash
#!/bin/bash
# auto-scale-tests.sh - Dynamically scale test workers based on queue

PROJECT="test-suite"
MIN_WORKERS=2
MAX_WORKERS=10
QUEUE_FILE="/tmp/test-queue"

# Function to get queue size
get_queue_size() {
    if [ -f "$QUEUE_FILE" ]; then
        wc -l < "$QUEUE_FILE"
    else
        echo 0
    fi
}

# Function to spawn worker
spawn_worker() {
    local worker_id="$1"
    
    OUTPUT=$(gitpool claim $PROJECT --branch "auto-worker-$worker_id-$(date +%s)")
    WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
    WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)
    
    echo "🤖 Spawned worker $worker_id: $WORKTREE_ID"
    
    (
        cd "$WORKTREE_PATH"
        
        while true; do
            # Get next test from queue
            if [ -f "$QUEUE_FILE" ]; then
                test_case=$(head -n1 "$QUEUE_FILE" 2>/dev/null)
                if [ -n "$test_case" ]; then
                    # Remove from queue
                    tail -n +2 "$QUEUE_FILE" > "$QUEUE_FILE.tmp" && mv "$QUEUE_FILE.tmp" "$QUEUE_FILE"
                    
                    echo "🧪 Worker $worker_id running: $test_case"
                    npm test -- "$test_case"
                else
                    sleep 1
                fi
            else
                sleep 1
            fi
        done
        
        gitpool release $WORKTREE_ID
    ) &
    
    echo $! > "/tmp/worker-$worker_id.pid"
}

# Monitoring loop
active_workers=0

while true; do
    queue_size=$(get_queue_size)
    
    if [ $queue_size -gt 0 ] && [ $active_workers -lt $MAX_WORKERS ]; then
        # Scale up
        active_workers=$((active_workers + 1))
        spawn_worker $active_workers
    elif [ $queue_size -eq 0 ] && [ $active_workers -gt $MIN_WORKERS ]; then
        # Scale down
        if [ -f "/tmp/worker-$active_workers.pid" ]; then
            kill $(cat "/tmp/worker-$active_workers.pid") 2>/dev/null
            rm "/tmp/worker-$active_workers.pid"
        fi
        active_workers=$((active_workers - 1))
    fi
    
    sleep 5
done
```

## Tips & Best Practices

### Error Handling
```bash
# Always use proper cleanup
cleanup() {
    if [ -n "$WORKTREE_ID" ]; then
        gitpool release "$WORKTREE_ID"
    fi
}
trap cleanup EXIT ERR

# Handle gitpool failures gracefully
claim_worktree() {
    local project="$1"
    local branch="$2"
    local max_retries=3
    local retry=0
    
    while [ $retry -lt $max_retries ]; do
        if OUTPUT=$(gitpool claim "$project" --branch "$branch" 2>/dev/null); then
            export WORKTREE_ID=$(echo "$OUTPUT" | head -n1)
            export WORKTREE_PATH=$(echo "$OUTPUT" | tail -n1)
            return 0
        fi
        
        retry=$((retry + 1))
        echo "⚠️ Claim failed, retrying ($retry/$max_retries)..."
        sleep $((retry * 2))
    done
    
    echo "❌ Failed to claim worktree after $max_retries attempts"
    return 1
}
```

### Resource Management
```bash
# Monitor and cleanup stale worktrees
cleanup_stale_worktrees() {
    # List worktrees older than 1 hour
    gitpool list --format json | jq -r '
        .[] | select(.claimed_at != null) | 
        select((now - (.claimed_at | fromdateiso8601)) > 3600) |
        .id
    ' | while read -r worktree_id; do
        echo "🧹 Releasing stale worktree: $worktree_id"
        gitpool release "$worktree_id"
    done
}

# Set up periodic cleanup
(
    while true; do
        sleep 3600  # Every hour
        cleanup_stale_worktrees
    done
) &
```

These examples showcase GitPool's versatility across different development scenarios. Choose the patterns that best fit your workflow and customize them for your specific needs.

<style>
/* Examples page styling */
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

.language-bash,
.language-yaml,
.language-groovy {
  position: relative;
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

/* Add copy button styles */
pre {
  position: relative;
}

pre:hover::after {
  content: "📋";
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  background: rgba(255,255,255,0.1);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
}

/* Highlight important sections */
strong {
  color: #667eea;
}

/* Section dividers */
hr {
  border: none;
  height: 2px;
  background: linear-gradient(90deg, #667eea, transparent);
  margin: 3rem 0;
}

/* Tips callouts */
blockquote {
  border-left: 4px solid #28a745;
  background: #d4edda;
  padding: 1rem;
  margin: 1rem 0;
  border-radius: 0 8px 8px 0;
}

/* Code spans */
code {
  background: #f8f9fa;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  font-size: 0.9em;
  color: #e83e8c;
}
</style>