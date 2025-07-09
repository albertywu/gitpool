package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/uber/treefarm/config"
	"github.com/uber/treefarm/models"
)

type Store struct {
	db *sql.DB
}

func NewStore() (*Store, error) {
	return NewStoreWithPath(config.GetWorktreeDir())
}

func NewStoreWithPath(worktreeDir string) (*Store, error) {
	dbPath := filepath.Join(worktreeDir, "treefarm.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS repositories (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			path TEXT NOT NULL,
			max_worktrees INTEGER NOT NULL,
			default_branch TEXT NOT NULL,
			fetch_interval INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS worktrees (
			id TEXT PRIMARY KEY,
			repo_id TEXT NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			status TEXT NOT NULL,
			leased_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS reconciler_runs (
			id TEXT PRIMARY KEY,
			run_time TIMESTAMP NOT NULL,
			created INTEGER NOT NULL,
			cleaned INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_worktrees_repo_id ON worktrees(repo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_worktrees_status ON worktrees(status)`,
		// Add last_fetch_time column to repositories table (safe if column already exists)
		`ALTER TABLE repositories ADD COLUMN last_fetch_time TIMESTAMP`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			// Ignore "duplicate column name" error for ALTER TABLE statements
			if query[:11] == "ALTER TABLE" && (err.Error() == "duplicate column name: last_fetch_time" ||
				err.Error() == "SQL logic error: duplicate column name: last_fetch_time") {
				continue
			}
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// Repository methods
func (s *Store) CreateRepository(repo *models.Repository) error {
	query := `INSERT INTO repositories (id, name, path, max_worktrees, default_branch, fetch_interval, last_fetch_time, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, repo.ID.String(), repo.Name, repo.Path, repo.MaxWorktrees,
		repo.DefaultBranch, repo.FetchInterval, repo.LastFetchTime, repo.CreatedAt)
	return err
}

func (s *Store) GetRepository(name string) (*models.Repository, error) {
	query := `SELECT id, name, path, max_worktrees, default_branch, fetch_interval, last_fetch_time, created_at 
			  FROM repositories WHERE name = ?`
	row := s.db.QueryRow(query, name)

	var repo models.Repository
	var idStr string
	err := row.Scan(&idStr, &repo.Name, &repo.Path, &repo.MaxWorktrees,
		&repo.DefaultBranch, &repo.FetchInterval, &repo.LastFetchTime, &repo.CreatedAt)
	if err != nil {
		return nil, err
	}

	repo.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

func (s *Store) ListRepositories() ([]*models.Repository, error) {
	query := `SELECT id, name, path, max_worktrees, default_branch, fetch_interval, last_fetch_time, created_at 
			  FROM repositories ORDER BY name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*models.Repository
	for rows.Next() {
		var repo models.Repository
		var idStr string
		err := rows.Scan(&idStr, &repo.Name, &repo.Path, &repo.MaxWorktrees,
			&repo.DefaultBranch, &repo.FetchInterval, &repo.LastFetchTime, &repo.CreatedAt)
		if err != nil {
			return nil, err
		}

		repo.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}

		repos = append(repos, &repo)
	}

	return repos, rows.Err()
}

func (s *Store) DeleteRepository(name string) error {
	query := `DELETE FROM repositories WHERE name = ?`
	_, err := s.db.Exec(query, name)
	return err
}

func (s *Store) UpdateRepositoryLastFetch(name string, lastFetchTime time.Time) error {
	query := `UPDATE repositories SET last_fetch_time = ? WHERE name = ?`
	_, err := s.db.Exec(query, lastFetchTime, name)
	return err
}

// Worktree methods
func (s *Store) CreateWorktree(worktree *models.Worktree) error {
	query := `INSERT INTO worktrees (id, repo_id, name, path, status, leased_at, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, worktree.ID.String(), worktree.RepoID.String(), worktree.Name,
		worktree.Path, worktree.Status, worktree.LeasedAt, worktree.CreatedAt)
	return err
}

func (s *Store) GetWorktree(id string) (*models.Worktree, error) {
	query := `SELECT id, repo_id, name, path, status, leased_at, created_at 
			  FROM worktrees WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var worktree models.Worktree
	var idStr, repoIDStr string
	err := row.Scan(&idStr, &repoIDStr, &worktree.Name, &worktree.Path,
		&worktree.Status, &worktree.LeasedAt, &worktree.CreatedAt)
	if err != nil {
		return nil, err
	}

	worktree.ID, _ = uuid.Parse(idStr)
	worktree.RepoID, _ = uuid.Parse(repoIDStr)

	return &worktree, nil
}

func (s *Store) GetWorktreeByName(name string) (*models.Worktree, error) {
	query := `SELECT id, repo_id, name, path, status, leased_at, created_at 
			  FROM worktrees WHERE name = ?`
	row := s.db.QueryRow(query, name)

	var worktree models.Worktree
	var idStr, repoIDStr string
	err := row.Scan(&idStr, &repoIDStr, &worktree.Name, &worktree.Path,
		&worktree.Status, &worktree.LeasedAt, &worktree.CreatedAt)
	if err != nil {
		return nil, err
	}

	worktree.ID, _ = uuid.Parse(idStr)
	worktree.RepoID, _ = uuid.Parse(repoIDStr)

	return &worktree, nil
}

func (s *Store) ListWorktreesByRepo(repoID uuid.UUID) ([]*models.Worktree, error) {
	query := `SELECT id, repo_id, name, path, status, leased_at, created_at 
			  FROM worktrees WHERE repo_id = ?`
	rows, err := s.db.Query(query, repoID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanWorktrees(rows)
}

func (s *Store) ListIdleWorktreesByRepo(repoID uuid.UUID) ([]*models.Worktree, error) {
	query := `SELECT id, repo_id, name, path, status, leased_at, created_at 
			  FROM worktrees WHERE repo_id = ? AND status = ?`
	rows, err := s.db.Query(query, repoID.String(), models.WorktreeStatusIdle)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanWorktrees(rows)
}

func (s *Store) UpdateWorktreeStatus(id string, status models.WorktreeStatus, leasedAt *time.Time) error {
	query := `UPDATE worktrees SET status = ?, leased_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, status, leasedAt, id)
	return err
}

func (s *Store) DeleteWorktree(id string) error {
	query := `DELETE FROM worktrees WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *Store) CountWorktreesByStatus(repoID uuid.UUID) (map[models.WorktreeStatus]int, error) {
	query := `SELECT status, COUNT(*) FROM worktrees WHERE repo_id = ? GROUP BY status`
	rows, err := s.db.Query(query, repoID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[models.WorktreeStatus]int)
	for rows.Next() {
		var status models.WorktreeStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}

	return counts, rows.Err()
}

// Reconciler methods
func (s *Store) CreateReconcilerRun(run *models.ReconcilerRun) error {
	query := `INSERT INTO reconciler_runs (id, run_time, created, cleaned) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, run.ID.String(), run.RunTime, run.Created, run.Cleaned)
	return err
}

func (s *Store) GetLastReconcilerRun() (*models.ReconcilerRun, error) {
	query := `SELECT id, run_time, created, cleaned FROM reconciler_runs ORDER BY run_time DESC LIMIT 1`
	row := s.db.QueryRow(query)

	var run models.ReconcilerRun
	var idStr string
	err := row.Scan(&idStr, &run.RunTime, &run.Created, &run.Cleaned)
	if err != nil {
		return nil, err
	}

	run.ID, _ = uuid.Parse(idStr)
	return &run, nil
}

// Helper methods
func (s *Store) scanWorktrees(rows *sql.Rows) ([]*models.Worktree, error) {
	var worktrees []*models.Worktree
	for rows.Next() {
		var worktree models.Worktree
		var idStr, repoIDStr string
		err := rows.Scan(&idStr, &repoIDStr, &worktree.Name, &worktree.Path,
			&worktree.Status, &worktree.LeasedAt, &worktree.CreatedAt)
		if err != nil {
			return nil, err
		}

		worktree.ID, _ = uuid.Parse(idStr)
		worktree.RepoID, _ = uuid.Parse(repoIDStr)

		worktrees = append(worktrees, &worktree)
	}

	return worktrees, rows.Err()
}
