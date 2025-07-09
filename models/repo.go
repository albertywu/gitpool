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
	DefaultBranch string     `db:"default_branch"`
	FetchInterval int        `db:"fetch_interval"` // minutes
	LastFetchTime *time.Time `db:"last_fetch_time"`
	CreatedAt     time.Time  `db:"created_at"`
}

func NewRepository(name, path, defaultBranch string, maxWorktrees, fetchInterval int) *Repository {
	return &Repository{
		ID:            uuid.New(),
		Name:          name,
		Path:          path,
		MaxWorktrees:  maxWorktrees,
		DefaultBranch: defaultBranch,
		FetchInterval: fetchInterval,
		LastFetchTime: nil, // No fetch has happened yet
		CreatedAt:     time.Now(),
	}
}
