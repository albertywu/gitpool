package daemon

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/albertywu/gitpool/config"
	"github.com/albertywu/gitpool/db"
	"github.com/albertywu/gitpool/models"
	"github.com/albertywu/gitpool/pool"
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

	for _, repo := range repos {
		// Get the fetch interval from config (defaults to 1h if not set)
		fetchInterval := r.config.GetRepoFetchInterval(repo.Name)

		// Check if it's time to fetch for this repo
		if repo.LastFetchTime == nil || time.Since(*repo.LastFetchTime) >= fetchInterval {
			log.Printf("[INFO] Processing repository '%s'", repo.Name)

			run, err := r.pool.ReconcileWorktrees(repo)
			if err != nil {
				log.Printf("[ERROR] Failed to reconcile worktrees for '%s': %v", repo.Name, err)
				continue
			}

			// Update last fetch time
			if err := r.store.UpdateRepositoryLastFetch(repo.Name, time.Now()); err != nil {
				log.Printf("[ERROR] Failed to update last fetch time for '%s': %v", repo.Name, err)
			}

			totalRun.Created += run.Created
			totalRun.Cleaned += run.Cleaned
		}
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
