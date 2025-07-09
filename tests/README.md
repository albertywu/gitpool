# GitPool Integration Tests

This directory contains integration tests for the gitpool CLI tool. These tests verify that all CLI commands work correctly together in a real environment.

## Test Coverage

The integration test suite covers all major CLI commands:

### Daemon Commands
- `gitpool daemon start` - Starting the background daemon
- `gitpool daemon status` - Checking daemon status

### Repository Management
- `gitpool repo add` - Adding repositories with various configurations
- `gitpool repo list` - Listing registered repositories
- `gitpool repo remove` - Removing repositories

### Worktree Operations
- `gitpool claim --repo <name>` - Claiming worktrees (returns both ID and path)
- `gitpool release <worktree-id>` - Releasing worktrees back to pool
- `gitpool pool status` - Checking pool status and statistics

### Full Workflow Tests
- Complete end-to-end workflows combining multiple commands
- Multiple repository management
- Concurrent worktree operations

## Running the Tests

### Quick Start
```bash
# Run all integration tests
make test-integration

# Run with verbose output
make test-integration-verbose

# Run directly with script
cd tests
./run_tests.sh

# Run with verbose output
cd tests
VERBOSE=true ./run_tests.sh
```

### Test Environment

The tests create a completely isolated environment:
- **Temporary directory**: All test files are created in `/tmp/gitpool-test-*`
- **Isolated config**: Uses temporary `.gitpool` directory
- **Test repositories**: Creates fresh git repositories for each test
- **Separate binary**: Builds gitpool binary in test directory
- **Clean cleanup**: Removes all temporary files after tests

### Test Structure

Each test:
1. Creates a fresh test environment
2. Builds the gitpool binary
3. Creates test git repositories
4. Runs CLI commands and verifies output
5. Cleans up all resources

## Test Details

### TestDaemonCommands
Tests daemon lifecycle management:
- Starting the daemon process
- Verifying daemon status reporting
- Proper daemon shutdown

### TestRepoCommands
Tests repository management:
- Adding repositories with different configurations
- Listing repositories and verifying output format
- Removing repositories and cleanup

### TestWorktreeCommands
Tests worktree operations:
- Claiming worktrees from repository pools (verifies both ID and path output)
- Checking pool status and statistics
- Releasing worktrees back to the pool
- Verifying pool state after operations

### TestFullWorkflow
Tests complete usage scenarios:
- Multiple repository management
- Concurrent worktree operations
- Pool capacity management
- End-to-end workflow validation

## Test Configuration

Tests use the following configuration:
- **Test timeout**: 300 seconds (5 minutes)
- **Daemon startup wait**: 2 seconds
- **Worktree creation wait**: 3 seconds
- **Repository max worktrees**: 2-4 (varies by test)
- **Default branch**: `main`

## Troubleshooting

### Common Issues

**Test timeouts**: If tests are timing out, increase `TEST_TIMEOUT` in `run_tests.sh`

**Port conflicts**: Tests use Unix sockets, so port conflicts shouldn't occur

**Permission errors**: Ensure the test script is executable: `chmod +x run_tests.sh`

**Stale processes**: Clean up with: `pkill -f "gitpool.*daemon"`

### Debug Mode

Run tests with verbose output to see detailed command output:
```bash
VERBOSE=true ./run_tests.sh
```

### Manual Testing

You can manually run individual test functions:
```bash
# Build test binary
go build -o integration_test integration_test.go

# Run specific test
./integration_test -test.run TestDaemonCommands
```

## Adding New Tests

To add new integration tests:

1. **Add test function**: Create a new test function following the pattern:
   ```go
   func TestNewFeature(t *testing.T) {
       tc := SetupTestContext(t)
       defer tc.TeardownTestContext()
       
       // Your test logic here
   }
   ```

2. **Register test**: Add to the `testing.Main` call in `main()`:
   ```go
   {"TestNewFeature", TestNewFeature},
   ```

3. **Follow patterns**: Use existing helper functions and patterns from other tests

## CI/CD Integration

The test suite is designed to work in CI/CD environments:
- Self-contained with no external dependencies
- Proper cleanup prevents resource leaks
- Exit codes indicate test success/failure
- Timeout protection prevents hanging builds

Example CI usage:
```bash
# In CI pipeline
make test-integration
```

## Performance Considerations

The integration tests:
- Create real git repositories (can be slow)
- Start actual daemon processes
- Perform real file system operations
- Typical runtime: 30-60 seconds

For faster feedback during development, consider running unit tests first:
```bash
make test-unit
``` 