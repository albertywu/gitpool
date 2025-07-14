package repo

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/albertywu/gitpool/internal/db"
	"github.com/albertywu/gitpool/internal/models"
)

type Manager struct {
	store     *db.Store
	validator *Validator
}

func NewManager(store *db.Store) *Manager {
	return &Manager{
		store:     store,
		validator: NewValidator(),
	}
}

func (m *Manager) AddRepository(name, path, baseBranch string, maxWorktrees int) (*models.Repository, error) {
	// Validate repository path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	if err := m.validator.ValidateRepository(absPath); err != nil {
		return nil, fmt.Errorf("repository validation failed: %w", err)
	}

	// Auto-detect base branch if not provided
	if baseBranch == "" {
		detected, err := m.validator.GetDefaultBranch(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to detect base branch: %w", err)
		}
		baseBranch = detected
		log.Printf("[INFO] Auto-detected base branch: %s", baseBranch)
	}

	// Validate base branch
	if err := m.validator.ValidateBranch(absPath, baseBranch); err != nil {
		return nil, fmt.Errorf("branch validation failed: %w", err)
	}

	// Check if repository already exists
	if _, err := m.store.GetRepository(name); err == nil {
		return nil, fmt.Errorf("repository '%s' already exists", name)
	}

	// Create repository record - no fetch interval, refresh is manual
	repo := models.NewRepository(name, absPath, baseBranch, maxWorktrees, 0)
	if err := m.store.CreateRepository(repo); err != nil {
		return nil, fmt.Errorf("failed to save repository: %w", err)
	}

	log.Printf("[INFO] Added repo '%s' at %s", name, absPath)
	log.Printf("[INFO] Max worktrees: %d, Base branch: %s",
		maxWorktrees, baseBranch)

	return repo, nil
}

func (m *Manager) ListRepositories() ([]*models.Repository, error) {
	return m.store.ListRepositories()
}

func (m *Manager) RemoveRepository(name string) error {
	// Get repository
	repo, err := m.store.GetRepository(name)
	if err != nil {
		return fmt.Errorf("repository '%s' not found", name)
	}

	// Get all worktrees for this repository
	worktrees, err := m.store.ListWorktreesByRepo(repo.ID)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Count in-use worktrees
	inUseCount := 0
	for _, wt := range worktrees {
		if wt.Status == models.WorktreeStatusInUse {
			inUseCount++
		}
	}

	if inUseCount > 0 {
		return fmt.Errorf("cannot remove repository with %d worktrees in use", inUseCount)
	}

	log.Printf("[WARN] Removing repo '%s'", name)

	// Delete worktree directories
	deletedCount := 0
	for _, wt := range worktrees {
		if wt.Status != models.WorktreeStatusInUse {
			if err := os.RemoveAll(wt.Path); err != nil {
				log.Printf("[ERROR] Failed to delete worktree directory %s: %v", wt.Path, err)
			} else {
				deletedCount++
			}
			// Delete from database regardless
			m.store.DeleteWorktree(wt.ID.String())
		}
	}

	// Delete repository record
	if err := m.store.DeleteRepository(name); err != nil {
		return fmt.Errorf("failed to delete repository record: %w", err)
	}

	log.Printf("[INFO] Deleted %d idle worktrees", deletedCount)
	log.Printf("[INFO] Repo '%s' removed successfully", name)

	return nil
}

func (m *Manager) GetRepository(name string) (*models.Repository, error) {
	return m.store.GetRepository(name)
}
