package pool

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/albertywu/gitpool/db"
	"github.com/albertywu/gitpool/models"
	"github.com/google/uuid"
)

type Pool struct {
	store     *db.Store
	allocator *Allocator
	mu        sync.Mutex
}

func NewPool(store *db.Store) *Pool {
	return &Pool{
		store:     store,
		allocator: NewAllocator(),
	}
}

func (p *Pool) ClaimWorktree(repoName string, branch string) (*models.Worktree, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Get repository
	repo, err := p.store.GetRepository(repoName)
	if err != nil {
		return nil, fmt.Errorf("repository '%s' not found", repoName)
	}

	// Check if branch is already in use
	inUse, err := p.store.IsBranchInUseForRepo(repo.ID, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to check branch availability: %w", err)
	}
	if inUse {
		return nil, fmt.Errorf("branch '%s' is already in use by another workspace in this repository", branch)
	}

	// Get idle worktrees
	idleWorktrees, err := p.store.ListIdleWorktreesByRepo(repo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	if len(idleWorktrees) == 0 {
		log.Printf("[INFO] No available worktrees for '%s'. Waiting...", repoName)

		// Trigger creation of new worktree if under capacity
		worktrees, _ := p.store.ListWorktreesByRepo(repo.ID)
		if len(worktrees) < repo.MaxWorktrees {
			log.Printf("[INFO] Triggering Reconciler to create a new worktree...")
			// In a real implementation, this would signal the reconciler
			// For now, we'll create one directly
			if err := p.createWorktree(repo); err != nil {
				return nil, fmt.Errorf("failed to create worktree: %w", err)
			}

			// Get the newly created worktree
			idleWorktrees, err = p.store.ListIdleWorktreesByRepo(repo.ID)
			if err != nil || len(idleWorktrees) == 0 {
				return nil, fmt.Errorf("failed to get newly created worktree")
			}
		} else {
			return nil, fmt.Errorf("no available worktrees and pool is at capacity")
		}
	}

	// Claim the first idle worktree
	worktree := idleWorktrees[0]
	claimedWorktree, err := p.allocator.ClaimWorktree(worktree)
	if err != nil {
		return nil, fmt.Errorf("failed to claim worktree: %w", err)
	}

	// Set the branch on the claimed worktree
	claimedWorktree.Branch = &branch

	// Update database with status and branch
	if err := p.store.UpdateWorktreeStatusAndBranch(claimedWorktree.ID.String(),
		claimedWorktree.Status, claimedWorktree.LeasedAt, claimedWorktree.Branch); err != nil {
		return nil, fmt.Errorf("failed to update worktree status: %w", err)
	}

	return claimedWorktree, nil
}

func (p *Pool) ReleaseWorktree(worktreeID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Get worktree by name or ID
	worktree, err := p.store.GetWorktreeByName(worktreeID)
	if err != nil {
		// Try by ID
		worktree, err = p.store.GetWorktree(worktreeID)
		if err != nil {
			return fmt.Errorf("worktree '%s' not found", worktreeID)
		}
	}

	log.Printf("[INFO] Releasing worktree '%s'", worktree.Name)
	log.Printf("[INFO] Cleaning worktree: git reset --hard, git clean -fdx")

	// Release worktree
	releasedWorktree, err := p.allocator.ReleaseWorktree(worktree)
	if err != nil {
		// Mark as corrupt if cleanup failed
		p.store.UpdateWorktreeStatus(worktree.ID.String(), models.WorktreeStatusCorrupt, nil)
		log.Printf("[INFO] Scheduling deletion and replacement of corrupted worktree")
		return fmt.Errorf("failed to release worktree: %w", err)
	}

	// Clear the branch when releasing
	releasedWorktree.Branch = nil

	// Update database
	if err := p.store.UpdateWorktreeStatusAndBranch(releasedWorktree.ID.String(),
		releasedWorktree.Status, releasedWorktree.LeasedAt, releasedWorktree.Branch); err != nil {
		return fmt.Errorf("failed to update worktree status: %w", err)
	}

	log.Printf("[INFO] Worktree returned to pool")
	return nil
}

func (p *Pool) GetPoolStatus(repoName string) ([]*models.PoolStatus, error) {
	var repos []*models.Repository
	var err error

	if repoName != "" {
		repo, err := p.store.GetRepository(repoName)
		if err != nil {
			return nil, fmt.Errorf("repository '%s' not found", repoName)
		}
		repos = []*models.Repository{repo}
	} else {
		repos, err = p.store.ListRepositories()
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}
	}

	var statuses []*models.PoolStatus
	for _, repo := range repos {
		counts, err := p.store.CountWorktreesByStatus(repo.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count worktrees: %w", err)
		}

		total := 0
		inUse := counts[models.WorktreeStatusInUse]
		idle := counts[models.WorktreeStatusIdle]
		for _, count := range counts {
			total += count
		}

		// Get last fetch time (would be tracked in real implementation)
		lastFetch := time.Now().Add(-time.Duration(repo.FetchInterval) * time.Minute)

		status := &models.PoolStatus{
			RepoName:  repo.Name,
			Total:     total,
			InUse:     inUse,
			Idle:      idle,
			Max:       repo.MaxWorktrees,
			LastFetch: &lastFetch,
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (p *Pool) createWorktree(repo *models.Repository) error {
	worktree, err := p.allocator.CreateWorktree(repo)
	if err != nil {
		return err
	}

	return p.store.CreateWorktree(worktree)
}

func (p *Pool) CreateInitialWorktrees(repo *models.Repository, count int) error {
	log.Printf("[INFO] Creating initial worktrees...")

	created := 0
	for i := 0; i < count && i < repo.MaxWorktrees; i++ {
		if err := p.createWorktree(repo); err != nil {
			log.Printf("[ERROR] Failed to create worktree: %v", err)
			continue
		}
		created++
	}

	if created > 0 {
		log.Printf("[INFO] Created %d worktree(s)", created)
	}

	return nil
}

func (p *Pool) ReconcileWorktrees(repo *models.Repository) (*models.ReconcilerRun, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	run := &models.ReconcilerRun{
		ID:      uuid.New(),
		RunTime: time.Now(),
		Created: 0,
		Cleaned: 0,
	}

	// Get all worktrees
	worktrees, err := p.store.ListWorktreesByRepo(repo.ID)
	if err != nil {
		return run, fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Count current status
	idleCount := 0
	corruptCount := 0
	for _, wt := range worktrees {
		switch wt.Status {
		case models.WorktreeStatusIdle:
			idleCount++
		case models.WorktreeStatusCorrupt:
			corruptCount++
		}
	}

	// Clean up corrupt worktrees
	for _, wt := range worktrees {
		if wt.Status == models.WorktreeStatusCorrupt {
			if err := p.allocator.DeleteWorktree(repo, wt); err != nil {
				log.Printf("[ERROR] Failed to delete corrupt worktree %s: %v", wt.Name, err)
			} else {
				p.store.DeleteWorktree(wt.ID.String())
				run.Cleaned++
			}
		}
	}

	// Create new worktrees if under capacity
	currentCount := len(worktrees) - run.Cleaned
	targetCount := repo.MaxWorktrees

	if currentCount < targetCount {
		toCreate := targetCount - currentCount

		for i := 0; i < toCreate; i++ {
			if err := p.createWorktree(repo); err != nil {
				log.Printf("[ERROR] Failed to create worktree: %v", err)
			} else {
				run.Created++
			}
		}
	}

	// Fetch updates for repository
	if err := p.allocator.FetchRepository(repo); err != nil {
		log.Printf("[ERROR] Failed to fetch repository updates: %v", err)
	}

	// Update idle worktrees
	idleWorktrees, _ := p.store.ListIdleWorktreesByRepo(repo.ID)
	for _, wt := range idleWorktrees {
		if err := p.allocator.UpdateWorktree(repo, wt); err != nil {
			log.Printf("[ERROR] Failed to update worktree %s: %v", wt.Name, err)
		}
	}

	return run, nil
}
