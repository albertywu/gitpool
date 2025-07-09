package pool

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/albertywu/gitpool/config"
	"github.com/albertywu/gitpool/models"
)

type Allocator struct {
}

func NewAllocator() *Allocator {
	return &Allocator{}
}

func (a *Allocator) CreateWorktree(repo *models.Repository) (*models.Worktree, error) {
	// Generate unique worktree name
	worktreeID := uuid.New()
	worktreeName := fmt.Sprintf("%s-%s", repo.Name, worktreeID.String())

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

	// Pull latest changes
	cmd := exec.Command("git", "-C", worktree.Path, "pull", "origin", repo.DefaultBranch)
	if err := cmd.Run(); err != nil {
		// If pull fails, try to reset to origin
		cmd = exec.Command("git", "-C", worktree.Path, "reset", "--hard", fmt.Sprintf("origin/%s", repo.DefaultBranch))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update worktree: %w", err)
		}
	}

	return nil
}

func (a *Allocator) ClaimWorktree(worktree *models.Worktree) (*models.Worktree, error) {
	if worktree.Status != models.WorktreeStatusIdle {
		return nil, fmt.Errorf("worktree is not idle")
	}

	now := time.Now()
	worktree.Status = models.WorktreeStatusInUse
	worktree.LeasedAt = &now

	return worktree, nil
}

func (a *Allocator) ReleaseWorktree(worktree *models.Worktree) (*models.Worktree, error) {
	if worktree.Status != models.WorktreeStatusInUse {
		return nil, fmt.Errorf("worktree is not in use")
	}

	// Try to clean the worktree
	if err := a.CleanWorktree(worktree); err != nil {
		log.Printf("[ERROR] Failed to clean worktree '%s': %v", worktree.Name, err)
		worktree.Status = models.WorktreeStatusCorrupt
		return worktree, fmt.Errorf("worktree cleanup failed")
	}

	worktree.Status = models.WorktreeStatusIdle
	worktree.LeasedAt = nil

	return worktree, nil
}
