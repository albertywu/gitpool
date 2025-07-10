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

	// Distribute initial fetch times for repos without LastFetchTime
	r.distributeInitialFetchTimes(repos)

	// Group repos by their fetch interval to prevent clustering
	reposByInterval := make(map[time.Duration][]*models.Repository)
	for _, repo := range repos {
		interval := r.config.GetRepoFetchInterval(repo.Name)
		reposByInterval[interval] = append(reposByInterval[interval], repo)
	}

	// Process repos, adding jitter within each interval group
	for interval, intervalRepos := range reposByInterval {
		for i, repo := range intervalRepos {
			// Check if it's time to fetch for this repo
			if repo.LastFetchTime == nil || time.Since(*repo.LastFetchTime) >= interval {
				// Add a small jitter to prevent repos with same interval from fetching together
				// Jitter is up to 10% of the reconciler interval, distributed across repos
				jitter := time.Duration(0)
				if len(intervalRepos) > 1 {
					maxJitter := r.interval / 10 // 10% of reconciler interval
					jitter = maxJitter * time.Duration(i) / time.Duration(len(intervalRepos))
					time.Sleep(jitter)
				}

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

// distributeInitialFetchTimes assigns staggered fetch times to repositories without LastFetchTime
// This prevents all repos from fetching at the same time when the daemon starts
func (r *Reconciler) distributeInitialFetchTimes(repos []*models.Repository) {
	// Count repos that need initial fetch times
	var reposNeedingInit []*models.Repository
	for _, repo := range repos {
		if repo.LastFetchTime == nil {
			reposNeedingInit = append(reposNeedingInit, repo)
		}
	}

	if len(reposNeedingInit) == 0 {
		return
	}

	log.Printf("[INFO] Distributing initial fetch times for %d repositories", len(reposNeedingInit))

	// For each repo, assign a fetch time distributed across its fetch interval
	now := time.Now()
	for i, repo := range reposNeedingInit {
		// Get the fetch interval for this repo
		fetchInterval := r.config.GetRepoFetchInterval(repo.Name)
		
		// Calculate offset: distribute repos evenly across the interval
		// For example, if we have 3 repos with 1h interval:
		// - Repo 1: now - 0h (fetches immediately)
		// - Repo 2: now - 20m (fetches in 40m)
		// - Repo 3: now - 40m (fetches in 20m)
		offset := fetchInterval * time.Duration(i) / time.Duration(len(reposNeedingInit))
		initialFetchTime := now.Add(-offset)
		
		// Update the repo's last fetch time
		if err := r.store.UpdateRepositoryLastFetch(repo.Name, initialFetchTime); err != nil {
			log.Printf("[ERROR] Failed to set initial fetch time for '%s': %v", repo.Name, err)
		} else {
			// Update the in-memory repo object so it's used in this reconcile run
			repo.LastFetchTime = &initialFetchTime
			nextFetch := fetchInterval - offset
			log.Printf("[INFO] Repository '%s' scheduled for initial fetch in %v", repo.Name, nextFetch)
		}
	}
}
