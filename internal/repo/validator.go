package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateRepository(path string) error {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	// Check if it's a git repository
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return fmt.Errorf("not a git repository (no .git directory)")
	}

	// Check if it's a bare repository
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-bare-repository")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check repository type: %w", err)
	}

	if string(output) == "true\n" {
		return fmt.Errorf("bare repositories are not supported")
	}

	// Verify we can list branches
	cmd = exec.Command("git", "-C", path, "branch", "-r")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	return nil
}

func (v *Validator) ValidateBranch(repoPath, branch string) error {
	// Check if branch exists
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--verify", branch)
	if err := cmd.Run(); err != nil {
		// Try as remote branch
		cmd = exec.Command("git", "-C", repoPath, "rev-parse", "--verify", "origin/"+branch)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("branch '%s' does not exist", branch)
		}
	}
	return nil
}
