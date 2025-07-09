package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestContext holds test configuration and state
type TestContext struct {
	TreefarmBinary string
	TestDir        string
	ConfigDir      string
	WorktreeDir    string
	TestRepo       string
	SocketPath     string
	DaemonCmd      *exec.Cmd
	t              *testing.T
}

// SetupTestContext creates a temporary test environment
func SetupTestContext(t *testing.T) *TestContext {
	// Create temporary test directory with short name to avoid Unix socket path length limits
	testDir, err := os.MkdirTemp("", "tf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Set up paths
	configDir := filepath.Join(testDir, ".treefarm")
	worktreeDir := filepath.Join(configDir, "worktrees")
	testRepo := filepath.Join(testDir, "test-repo")
	socketPath := filepath.Join(testDir, "daemon.sock")

	// Create directories
	os.MkdirAll(configDir, 0755)
	os.MkdirAll(worktreeDir, 0755)

	// Build treefarm binary
	treefarmBinary := filepath.Join(testDir, "treefarm")

	// Change to parent directory to build
	pwd, _ := os.Getwd()
	parentDir := filepath.Dir(pwd)

	cmd := exec.Command("go", "build", "-o", treefarmBinary, "./cmd/main.go")
	cmd.Dir = parentDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build treefarm: %v", err)
	}

	// Create test git repository
	createTestRepo(t, testRepo)

	return &TestContext{
		TreefarmBinary: treefarmBinary,
		TestDir:        testDir,
		ConfigDir:      configDir,
		WorktreeDir:    worktreeDir,
		TestRepo:       testRepo,
		SocketPath:     socketPath,
		t:              t,
	}
}

// TeardownTestContext cleans up the test environment
func (tc *TestContext) TeardownTestContext() {
	// Stop daemon if running
	tc.StopDaemon()

	// Clean up test directory
	os.RemoveAll(tc.TestDir)
}

// createTestRepo creates a test git repository with some commits
func createTestRepo(t *testing.T, repoPath string) {
	// Initialize git repo
	cmd := exec.Command("git", "init", repoPath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Set up git config
	cmd = exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com")
	cmd.Run()
	cmd = exec.Command("git", "-C", repoPath, "config", "user.name", "Test User")
	cmd.Run()

	// Create initial commit
	readmePath := filepath.Join(repoPath, "README.md")
	os.WriteFile(readmePath, []byte("# Test Repository\n"), 0644)

	cmd = exec.Command("git", "-C", repoPath, "add", "README.md")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add README: %v", err)
	}

	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create main branch (for compatibility)
	cmd = exec.Command("git", "-C", repoPath, "checkout", "-b", "main")
	cmd.Run()
}

// RunTreefarmCommand runs a treefarm command with the test environment
func (tc *TestContext) RunTreefarmCommand(args ...string) (string, error) {
	// Set HOME to our test directory so config is found in the right place
	cmd := exec.Command(tc.TreefarmBinary, args...)
	cmd.Env = append(os.Environ(),
		"HOME="+tc.TestDir,
		"TREEFARM_CONFIG_DIR="+tc.ConfigDir,
		"TREEFARM_WORKTREE_DIR="+tc.WorktreeDir,
		"TREEFARM_SOCKET_PATH="+tc.SocketPath)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// StartDaemon starts the treefarm daemon for testing
func (tc *TestContext) StartDaemon() error {
	// Start daemon as a subprocess (non-blocking)
	cmd := exec.Command(tc.TreefarmBinary, "daemon", "start",
		"--config-dir", tc.ConfigDir,
		"--worktree-dir", tc.WorktreeDir,
		"--socket-path", tc.SocketPath)
	cmd.Env = append(os.Environ(),
		"HOME="+tc.TestDir,
		"TREEFARM_CONFIG_DIR="+tc.ConfigDir,
		"TREEFARM_WORKTREE_DIR="+tc.WorktreeDir,
		"TREEFARM_SOCKET_PATH="+tc.SocketPath)

	// Create pipes to capture output for debugging
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Start the daemon process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon process: %v", err)
	}

	tc.DaemonCmd = cmd

	// Read stderr in background for debugging
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				break
			}
			if n > 0 {
				tc.t.Logf("Daemon stderr: %s", string(buf[:n]))
			}
		}
	}()

	// Wait for daemon to start and socket to be available
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)

		// Check if daemon process is still running
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			return fmt.Errorf("daemon process exited unexpectedly")
		}

		// Try to connect to daemon
		output, err := tc.RunTreefarmCommand("daemon", "status")
		if err == nil {
			return nil // Success!
		}

		tc.t.Logf("Daemon not ready yet (attempt %d/10): %v, output: %s", i+1, err, output)
	}

	return fmt.Errorf("daemon failed to start within timeout")
}

// StopDaemon stops the treefarm daemon
func (tc *TestContext) StopDaemon() error {
	if tc.DaemonCmd == nil {
		return nil
	}

	// Try to stop daemon gracefully first
	tc.RunTreefarmCommand("daemon", "stop")

	// Give it a moment to shutdown gracefully
	time.Sleep(1 * time.Second)

	// If still running, kill the process
	if tc.DaemonCmd.ProcessState == nil || !tc.DaemonCmd.ProcessState.Exited() {
		if err := tc.DaemonCmd.Process.Kill(); err != nil {
			tc.t.Logf("Failed to kill daemon process: %v", err)
		}
		tc.DaemonCmd.Wait() // Wait for process to exit
	}

	tc.DaemonCmd = nil
	return nil
}

// TestDaemonCommands tests daemon start and status commands
func TestDaemonCommands(t *testing.T) {
	tc := SetupTestContext(t)
	defer tc.TeardownTestContext()

	t.Run("daemon start", func(t *testing.T) {
		err := tc.StartDaemon()
		if err != nil {
			t.Fatalf("Failed to start daemon: %v", err)
		}
	})

	t.Run("daemon status", func(t *testing.T) {
		output, err := tc.RunTreefarmCommand("daemon", "status")
		if err != nil {
			t.Fatalf("Daemon status command failed: %v", err)
		}

		if !strings.Contains(output, "Daemon is running") {
			t.Errorf("Expected daemon to be running, got: %s", output)
		}
	})
}

// TestRepoCommands tests repository management commands
func TestRepoCommands(t *testing.T) {
	tc := SetupTestContext(t)
	defer tc.TeardownTestContext()

	// Start daemon
	if err := tc.StartDaemon(); err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}

	t.Run("repo add", func(t *testing.T) {
		output, err := tc.RunTreefarmCommand("repo", "add", "test-repo", tc.TestRepo, "--max", "4", "--default-branch", "main")
		if err != nil {
			t.Fatalf("Failed to add repo: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(output, "Repository added successfully") {
			t.Errorf("Expected success message, got: %s", output)
		}
	})

	t.Run("repo list", func(t *testing.T) {
		output, err := tc.RunTreefarmCommand("repo", "list")
		if err != nil {
			t.Fatalf("Failed to list repos: %v", err)
		}

		if !strings.Contains(output, "test-repo") {
			t.Errorf("Expected test-repo in list, got: %s", output)
		}

		if !strings.Contains(output, "main") {
			t.Errorf("Expected main branch in list, got: %s", output)
		}
	})

	t.Run("repo remove", func(t *testing.T) {
		output, err := tc.RunTreefarmCommand("repo", "remove", "test-repo")
		if err != nil {
			t.Fatalf("Failed to remove repo: %v", err)
		}

		if !strings.Contains(output, "Repository removed successfully") {
			t.Errorf("Expected success message, got: %s", output)
		}

		// Verify repo is removed
		output, err = tc.RunTreefarmCommand("repo", "list")
		if err != nil {
			t.Fatalf("Failed to list repos after removal: %v", err)
		}

		if strings.Contains(output, "test-repo") {
			t.Errorf("Expected test-repo to be removed, but still found in: %s", output)
		}
	})
}

// TestWorktreeCommands tests worktree operations
func TestWorktreeCommands(t *testing.T) {
	tc := SetupTestContext(t)
	defer tc.TeardownTestContext()

	// Start daemon
	if err := tc.StartDaemon(); err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}

	// Add repository
	_, err := tc.RunTreefarmCommand("repo", "add", "test-repo", tc.TestRepo, "--max", "2", "--default-branch", "main")
	if err != nil {
		t.Fatalf("Failed to add repo: %v", err)
	}

	// Wait for worktrees to be created
	time.Sleep(3 * time.Second)

	var worktreeID string

	t.Run("claim worktree", func(t *testing.T) {
		output, err := tc.RunTreefarmCommand("claim", "--repo", "test-repo")
		if err != nil {
			t.Fatalf("Failed to claim worktree: %v\nOutput: %s", err, output)
		}

		output = strings.TrimSpace(output)
		if !strings.HasPrefix(output, "test-repo-") {
			t.Errorf("Expected worktree name to start with 'test-repo-', got: %s", output)
		}

		worktreeID = output
	})

	t.Run("claim worktree with path", func(t *testing.T) {
		output, err := tc.RunTreefarmCommand("claim", "--repo", "test-repo", "--output-path")
		if err != nil {
			t.Fatalf("Failed to claim worktree with path: %v", err)
		}

		output = strings.TrimSpace(output)
		if !strings.Contains(output, "test-repo-") {
			t.Errorf("Expected path to contain 'test-repo-', got: %s", output)
		}

		// Verify the path exists
		if _, err := os.Stat(output); err != nil {
			t.Errorf("Expected worktree path to exist: %s", output)
		}
	})

	t.Run("pool status", func(t *testing.T) {
		output, err := tc.RunTreefarmCommand("pool", "status")
		if err != nil {
			t.Fatalf("Failed to get pool status: %v", err)
		}

		if !strings.Contains(output, "test-repo") {
			t.Errorf("Expected test-repo in pool status, got: %s", output)
		}

		// Should show some worktrees in use
		if !strings.Contains(output, "2") { // Total worktrees
			t.Errorf("Expected to see total worktrees in status, got: %s", output)
		}
	})

	t.Run("release worktree", func(t *testing.T) {
		if worktreeID == "" {
			t.Skip("No worktree ID to release")
		}

		output, err := tc.RunTreefarmCommand("release", worktreeID)
		if err != nil {
			t.Fatalf("Failed to release worktree: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(output, "Worktree released successfully") {
			t.Errorf("Expected success message, got: %s", output)
		}
	})

	t.Run("pool status after release", func(t *testing.T) {
		output, err := tc.RunTreefarmCommand("pool", "status")
		if err != nil {
			t.Fatalf("Failed to get pool status: %v", err)
		}

		if !strings.Contains(output, "test-repo") {
			t.Errorf("Expected test-repo in pool status, got: %s", output)
		}
	})
}

// TestFullWorkflow tests a complete workflow
func TestFullWorkflow(t *testing.T) {
	tc := SetupTestContext(t)
	defer tc.TeardownTestContext()

	// Start daemon
	if err := tc.StartDaemon(); err != nil {
		t.Fatalf("Failed to start daemon: %v", err)
	}

	// Add repository
	_, err := tc.RunTreefarmCommand("repo", "add", "workflow-repo", tc.TestRepo, "--max", "3", "--default-branch", "main")
	if err != nil {
		t.Fatalf("Failed to add repo: %v", err)
	}

	// Wait for worktrees to be created
	time.Sleep(3 * time.Second)

	// Claim multiple worktrees
	var worktreeIDs []string
	for i := 0; i < 2; i++ {
		output, err := tc.RunTreefarmCommand("claim", "--repo", "workflow-repo")
		if err != nil {
			t.Fatalf("Failed to claim worktree %d: %v", i, err)
		}
		worktreeIDs = append(worktreeIDs, strings.TrimSpace(output))
	}

	// Check pool status
	output, err := tc.RunTreefarmCommand("pool", "status")
	if err != nil {
		t.Fatalf("Failed to get pool status: %v", err)
	}

	if !strings.Contains(output, "workflow-repo") {
		t.Errorf("Expected workflow-repo in pool status, got: %s", output)
	}

	// Release worktrees
	for _, worktreeID := range worktreeIDs {
		_, err := tc.RunTreefarmCommand("release", worktreeID)
		if err != nil {
			t.Fatalf("Failed to release worktree %s: %v", worktreeID, err)
		}
	}

	// Final pool status check
	output, err = tc.RunTreefarmCommand("pool", "status")
	if err != nil {
		t.Fatalf("Failed to get final pool status: %v", err)
	}

	if !strings.Contains(output, "workflow-repo") {
		t.Errorf("Expected workflow-repo in final pool status, got: %s", output)
	}
}

// Note: This file uses standard Go testing. Run with: go test
