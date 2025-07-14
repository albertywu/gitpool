package models

import (
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	ID            uuid.UUID  `db:"id"`
	Name          string     `db:"name"`
	Path          string     `db:"path"`
	MaxWorktrees  int        `db:"max_worktrees"`
	BaseBranch    string     `db:"default_branch"` // Keep DB column name for compatibility
	FetchInterval int        `db:"fetch_interval"` // minutes
	LastFetchTime *time.Time `db:"last_fetch_time"`
	CreatedAt     time.Time  `db:"created_at"`
}

func NewRepository(name, path, baseBranch string, maxWorktrees, fetchInterval int) *Repository {
	return &Repository{
		ID:            uuid.New(),
		Name:          name,
		Path:          path,
		MaxWorktrees:  maxWorktrees,
		BaseBranch:    baseBranch,
		FetchInterval: fetchInterval,
		LastFetchTime: nil, // No fetch has happened yet
		CreatedAt:     time.Now(),
	}
}
