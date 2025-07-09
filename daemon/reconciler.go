package daemon

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/uber/treefarm/db"
	"github.com/uber/treefarm/models"
	"github.com/uber/treefarm/pool"
)

type Reconciler struct {
	store    *db.Store
	pool     *pool.Pool
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func NewReconciler(store *db.Store, pool *pool.Pool, interval time.Duration) *Reconciler {
	return &Reconciler{
		store:    store,
		pool:     pool,
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
		// Check if it's time to fetch for this repo
		lastFetch := time.Now().Add(-time.Duration(repo.FetchInterval) * time.Minute)
		if time.Since(lastFetch) >= time.Duration(repo.FetchInterval)*time.Minute {
			log.Printf("[INFO] Processing repository '%s'", repo.Name)

			run, err := r.pool.ReconcileWorktrees(repo)
			if err != nil {
				log.Printf("[ERROR] Failed to reconcile worktrees for '%s': %v", repo.Name, err)
				continue
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
