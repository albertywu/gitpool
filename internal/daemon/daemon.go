package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/albertywu/gitpool/internal/config"
	"github.com/albertywu/gitpool/internal/db"
	"github.com/albertywu/gitpool/internal/ipc"
	"github.com/albertywu/gitpool/internal/models"
	"github.com/albertywu/gitpool/internal/pool"
	"github.com/albertywu/gitpool/internal/repo"
)

type Daemon struct {
	config      *config.Config
	store       *db.Store
	repoManager *repo.Manager
	pool        *pool.Pool
	reconciler  *Reconciler
	server      *ipc.Server
	startTime   time.Time
	mu          sync.RWMutex
}

func New(cfg *config.Config) (*Daemon, error) {
	// Ensure work directory exists
	if err := cfg.EnsureWorktreeDir(); err != nil {
		return nil, fmt.Errorf("failed to create worktree directory: %w", err)
	}

	// Initialize database
	store, err := db.NewStoreWithPath(cfg.WorktreeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize components
	repoManager := repo.NewManager(store)
	worktreePool := pool.NewPool(store)
	reconciler := NewReconciler(store, worktreePool, cfg, cfg.ReconciliationInterval)

	d := &Daemon{
		config:      cfg,
		store:       store,
		repoManager: repoManager,
		pool:        worktreePool,
		reconciler:  reconciler,
		startTime:   time.Now(),
	}

	// Initialize IPC server
	server, err := ipc.NewServer(cfg.SocketPath, d)
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("failed to create IPC server: %w", err)
	}
	d.server = server

	return d, nil
}

func (d *Daemon) Start() error {
	log.Printf("[INFO] Starting treefarm daemon")
	log.Printf("[INFO] Using worktree directory: %s", d.config.WorktreeDir)
	log.Printf("[INFO] Global reconciliation interval: %s", d.config.ReconciliationInterval)
	log.Printf("[INFO] Listening on %s", d.config.SocketPath)

	// Start reconciler
	d.reconciler.Start()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start IPC server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := d.server.Serve(); err != nil {
			errCh <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigCh:
		log.Printf("[INFO] Received shutdown signal")
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	return d.Stop()
}

func (d *Daemon) Stop() error {
	log.Printf("[INFO] Stopping daemon...")

	// Stop reconciler
	d.reconciler.Stop()

	// Close server
	if err := d.server.Close(); err != nil {
		log.Printf("[ERROR] Failed to close server: %v", err)
	}

	// Close database
	if err := d.store.Close(); err != nil {
		log.Printf("[ERROR] Failed to close database: %v", err)
	}

	// Remove socket file
	os.Remove(d.config.SocketPath)

	log.Printf("[INFO] Daemon stopped")
	return nil
}

// IPC Handler implementations

func (d *Daemon) HandleRepoAdd(req ipc.RepoAddRequest) ipc.Response {
	d.mu.Lock()
	defer d.mu.Unlock()

	// No more fetch intervals - refresh is manual only
	repo, err := d.repoManager.AddRepository(req.Name, req.Path, req.BaseBranch,
		req.MaxWorktrees)
	if err != nil {
		return ipc.Response{Success: false, Error: err.Error()}
	}

	// Create initial worktrees up to the repository's max
	d.pool.CreateInitialWorktrees(repo, repo.MaxWorktrees)

	return ipc.Response{Success: true, Data: repo}
}

func (d *Daemon) HandleRepoList() ipc.Response {
	repos, err := d.repoManager.ListRepositories()
	if err != nil {
		return ipc.Response{Success: false, Error: err.Error()}
	}

	return ipc.Response{Success: true, Data: repos}
}

func (d *Daemon) HandleRepoRemove(name string) ipc.Response {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.repoManager.RemoveRepository(name); err != nil {
		return ipc.Response{Success: false, Error: err.Error()}
	}

	return ipc.Response{Success: true}
}

func (d *Daemon) HandleClaim(req ipc.ClaimRequest) ipc.Response {
	worktree, err := d.pool.ClaimWorktree(req.RepoName, req.Branch)
	if err != nil {
		return ipc.Response{Success: false, Error: err.Error()}
	}

	// Create the claim response with both ID and path
	claimResp := ipc.ClaimResponse{
		WorktreeID: worktree.Name, // Using Name as the identifier (e.g., "my-app-uuid")
		Path:       worktree.Path,
	}

	data, _ := json.Marshal(claimResp)
	return ipc.Response{Success: true, Data: json.RawMessage(data)}
}

func (d *Daemon) HandleRelease(req ipc.ReleaseRequest) ipc.Response {
	if err := d.pool.ReleaseWorktree(req.WorktreeID); err != nil {
		return ipc.Response{Success: false, Error: err.Error()}
	}

	return ipc.Response{Success: true}
}

func (d *Daemon) HandlePoolStatus(req ipc.PoolStatusRequest) ipc.Response {
	statuses, err := d.pool.GetPoolStatus(req.RepoName)
	if err != nil {
		return ipc.Response{Success: false, Error: err.Error()}
	}

	return ipc.Response{Success: true, Data: statuses}
}

func (d *Daemon) HandleDaemonStatus() ipc.Response {
	d.mu.RLock()
	defer d.mu.RUnlock()

	repos, _ := d.store.ListRepositories()
	lastRun, _ := d.store.GetLastReconcilerRun()

	var lastReconciler *time.Time
	if lastRun != nil {
		lastReconciler = &lastRun.RunTime
	}

	status := models.DaemonStatus{
		Running:        true,
		SocketPath:     d.config.SocketPath,
		LastReconciler: lastReconciler,
		Repositories:   len(repos),
	}

	// Convert to map for JSON response
	statusMap := map[string]interface{}{
		"running":         status.Running,
		"socket_path":     status.SocketPath,
		"last_reconciler": status.LastReconciler,
		"repositories":    status.Repositories,
		"uptime":          time.Since(d.startTime).String(),
	}

	data, _ := json.Marshal(statusMap)
	return ipc.Response{Success: true, Data: json.RawMessage(data)}
}

func (d *Daemon) HandleWorktreeList() ipc.Response {
	details, err := d.store.ListAllWorktreesWithRepos()
	if err != nil {
		return ipc.Response{Success: false, Error: err.Error()}
	}

	return ipc.Response{Success: true, Data: details}
}

func (d *Daemon) HandleRefresh(req ipc.RefreshRequest) ipc.Response {
	// Get repository
	repo, err := d.store.GetRepository(req.RepoName)
	if err != nil {
		return ipc.Response{Success: false, Error: fmt.Sprintf("repository '%s' not found", req.RepoName)}
	}

	// Manually trigger refresh for this repository
	log.Printf("[INFO] Manually refreshing repository '%s'", repo.Name)

	// Use the pool's ReconcileWorktrees which handles fetching and updating
	run, err := d.pool.ReconcileWorktrees(repo)
	if err != nil {
		return ipc.Response{Success: false, Error: fmt.Sprintf("refresh failed: %v", err)}
	}

	// Update last fetch time
	if err := d.store.UpdateRepositoryLastFetch(repo.Name, time.Now()); err != nil {
		log.Printf("[ERROR] Failed to update last fetch time: %v", err)
	}

	result := map[string]interface{}{
		"repository":        repo.Name,
		"worktrees_updated": run.Created,
		"worktrees_cleaned": run.Cleaned,
	}

	return ipc.Response{Success: true, Data: result}
}

func (d *Daemon) HandleShow(req ipc.ShowRequest) ipc.Response {
	// Try to get worktree by ID or name
	worktree, err := d.store.GetWorktree(req.WorktreeID)
	if err != nil {
		// Try by name
		worktree, err = d.store.GetWorktreeByName(req.WorktreeID)
		if err != nil {
			return ipc.Response{Success: false, Error: fmt.Sprintf("worktree '%s' not found", req.WorktreeID)}
		}
	}

	// Get repository info
	repo, err := d.store.GetRepositoryByID(worktree.RepoID)
	if err != nil {
		return ipc.Response{Success: false, Error: fmt.Sprintf("repository not found for worktree")}
	}

	// Create detail response
	detail := models.WorktreeDetail{
		Worktree:   worktree,
		Repository: repo,
	}

	return ipc.Response{Success: true, Data: detail}
}

func CheckDaemonRunning(socketPath string) bool {
	if _, err := os.Stat(socketPath); err != nil {
		return false
	}

	// Try to connect
	client := ipc.NewClient(socketPath)
	resp, err := client.DaemonStatus()
	if err != nil {
		return false
	}

	return resp.Success
}
