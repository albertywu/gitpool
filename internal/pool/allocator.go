package pool

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/albertywu/gitpool/internal/config"
	"github.com/albertywu/gitpool/internal/models"
	"github.com/google/uuid"
)

type Allocator struct {
}

func NewAllocator() *Allocator {
	return &Allocator{}
}

func (a *Allocator) CreateWorktree(repo *models.Repository) (*models.Worktree, error) {
	// Generate unique worktree name
	worktreeID := uuid.New()
	worktreeName := worktreeID.String()

	// Create repository subdirectory if needed
	repoWorkDir := filepath.Join(config.GetWorktreeDir(), repo.Name)
	if err := os.MkdirAll(repoWorkDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repo work directory: %w", err)
	}

	worktreePath := filepath.Join(repoWorkDir, worktreeName)

	// Create git worktree
	cmd := exec.Command("git", "-C", repo.Path, "worktree", "add", "--detach", worktreePath, repo.DefaultBranch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w\nOutput: %s", err, string(output))
	}

	// Create worktree model
	worktree := models.NewWorktree(repo.ID, worktreeName, worktreePath)

	log.Printf("[INFO] Created worktree: %s", worktreeName)

	return worktree, nil
}

func (a *Allocator) CleanWorktree(worktree *models.Worktree) error {
	// Reset to HEAD
	cmd := exec.Command("git", "-C", worktree.Path, "reset", "--hard", "HEAD")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reset worktree: %w", err)
	}

	// Clean untracked files
	cmd = exec.Command("git", "-C", worktree.Path, "clean", "-fdx")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clean worktree: %w", err)
	}

	return nil
}

func (a *Allocator) DeleteWorktree(repo *models.Repository, worktree *models.Worktree) error {
	// Remove git worktree
	cmd := exec.Command("git", "-C", repo.Path, "worktree", "remove", worktree.Path, "--force")
	if err := cmd.Run(); err != nil {
		// If worktree command fails, try to remove directory directly
		log.Printf("[WARN] Failed to remove worktree via git: %v", err)
	}

	// Remove directory if it still exists
	if err := os.RemoveAll(worktree.Path); err != nil {
		return fmt.Errorf("failed to remove worktree directory: %w", err)
	}

	// Prune worktree references
	cmd = exec.Command("git", "-C", repo.Path, "worktree", "prune")
	cmd.Run() // Ignore errors from prune

	return nil
}

func (a *Allocator) FetchRepository(repo *models.Repository) error {
	log.Printf("[INFO] Fetching updates for repository '%s'", repo.Name)

	cmd := exec.Command("git", "-C", repo.Path, "fetch", "--all", "--prune")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to fetch repository: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (a *Allocator) UpdateWorktree(repo *models.Repository, worktree *models.Worktree) error {
	// First clean the worktree
	if err := a.CleanWorktree(worktree); err != nil {
		return fmt.Errorf("failed to clean worktree: %w", err)
	}

	// Get the latest commit SHA for the default branch
	cmd := exec.Command("git", "-C", repo.Path, "rev-parse", fmt.Sprintf("origin/%s", repo.DefaultBranch))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get latest commit SHA: %w", err)
	}

	latestSHA := strings.TrimSpace(string(output))

	// Reset worktree to the latest commit (maintains detached HEAD state)
	cmd = exec.Command("git", "-C", worktree.Path, "reset", "--hard", latestSHA)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update worktree to %s: %w\nOutput: %s", latestSHA, err, string(output))
	}

	log.Printf("[INFO] Updated worktree %s to commit %s", worktree.Name, latestSHA[:7])

	return nil
}

func (a *Allocator) ClaimWorktree(worktree *models.Worktree, branch string) (*models.Worktree, error) {
	if worktree.Status != models.WorktreeStatusIdle {
		return nil, fmt.Errorf("worktree is not idle")
	}

	// First, fetch to ensure we have the latest branches
	cmd := exec.Command("git", "-C", worktree.Path, "fetch", "origin")
	if err := cmd.Run(); err != nil {
		log.Printf("[WARN] Failed to fetch before checkout: %v", err)
	}

	// Check if the branch exists locally or remotely
	var checkoutCmd *exec.Cmd

	// Try to checkout the branch (will create it from origin if it doesn't exist locally)
	checkoutCmd = exec.Command("git", "-C", worktree.Path, "checkout", "-B", branch, fmt.Sprintf("origin/%s", branch))
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		// If checkout from origin fails, try creating a new branch
		checkoutCmd = exec.Command("git", "-C", worktree.Path, "checkout", "-b", branch)
		if output2, err2 := checkoutCmd.CombinedOutput(); err2 != nil {
			return nil, fmt.Errorf("failed to checkout branch %s: %w\nOutput: %s\n%s", branch, err, string(output), string(output2))
		}
	}

	now := time.Now()
	worktree.Status = models.WorktreeStatusInUse
	worktree.LeasedAt = &now
	worktree.Branch = &branch

	log.Printf("[INFO] Claimed worktree %s with branch %s", worktree.Name, branch)

	return worktree, nil
}

func (a *Allocator) ReleaseWorktree(worktree *models.Worktree, repo *models.Repository) (*models.Worktree, error) {
	if worktree.Status != models.WorktreeStatusInUse {
		return nil, fmt.Errorf("worktree is not in use")
	}

	// Try to clean the worktree
	if err := a.CleanWorktree(worktree); err != nil {
		log.Printf("[ERROR] Failed to clean worktree '%s': %v", worktree.Name, err)
		worktree.Status = models.WorktreeStatusCorrupt
		return worktree, fmt.Errorf("worktree cleanup failed")
	}

	// Checkout back to detached HEAD at the default branch
	cmd := exec.Command("git", "-C", worktree.Path, "checkout", "--detach", fmt.Sprintf("origin/%s", repo.DefaultBranch))
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[WARN] Failed to detach HEAD: %v\nOutput: %s", err, string(output))
	}

	worktree.Status = models.WorktreeStatusIdle
	worktree.LeasedAt = nil
	worktree.Branch = nil

	return worktree, nil
}
