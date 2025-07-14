package daemon

import (
	"log"
	"sync"
	"time"

	"github.com/albertywu/gitpool/internal/config"
	"github.com/albertywu/gitpool/internal/db"
	"github.com/albertywu/gitpool/internal/models"
	"github.com/albertywu/gitpool/internal/pool"
	"github.com/google/uuid"
)

type Reconciler struct {
	store    *db.Store
	pool     *pool.Pool
	config   *config.Config
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func NewReconciler(store *db.Store, pool *pool.Pool, cfg *config.Config, interval time.Duration) *Reconciler {
	return &Reconciler{
		store:    store,
		pool:     pool,
		config:   cfg,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (r *Reconciler) Start() {
	r.wg.Add(1)
	go r.run()
}

func (r *Reconciler) Stop() {
	close(r.stopCh)
	r.wg.Wait()
}

func (r *Reconciler) run() {
	defer r.wg.Done()

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	// Run immediately on start
	r.reconcile()

	for {
		select {
		case <-ticker.C:
			r.reconcile()
		case <-r.stopCh:
			return
		}
	}
}

func (r *Reconciler) reconcile() {
	log.Printf("[INFO] Running reconciler...")

	totalRun := &models.ReconcilerRun{
		ID:      uuid.New(),
		RunTime: time.Now(),
		Created: 0,
		Cleaned: 0,
	}

	// Get all repositories
	repos, err := r.store.ListRepositories()
	if err != nil {
		log.Printf("[ERROR] Failed to list repositories: %v", err)
		return
	}

	// Process each repository - only maintain worktree pool size and clean corrupt worktrees
	// No automatic fetching - users must use 'gitpool refresh' command
	for _, repo := range repos {
		log.Printf("[INFO] Maintaining worktree pool for repository '%s'", repo.Name)

		// Only reconcile worktree pool (create/delete), don't fetch
		run, err := r.pool.MaintainWorktreePool(repo)
		if err != nil {
			log.Printf("[ERROR] Failed to maintain worktree pool for '%s': %v", repo.Name, err)
			continue
		}

		totalRun.Created += run.Created
		totalRun.Cleaned += run.Cleaned
	}

	// Save reconciler run
	if err := r.store.CreateReconcilerRun(totalRun); err != nil {
		log.Printf("[ERROR] Failed to save reconciler run: %v", err)
	}

	if totalRun.Created > 0 || totalRun.Cleaned > 0 {
		log.Printf("[INFO] Reconciler completed: created=%d, cleaned=%d",
			totalRun.Created, totalRun.Cleaned)
	}
}

func (r *Reconciler) TriggerReconcile() {
	go r.reconcile()
}
