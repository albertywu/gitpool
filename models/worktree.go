package models

import (
	"time"

	"github.com/google/uuid"
)

type WorktreeStatus string

const (
	WorktreeStatusIdle    WorktreeStatus = "idle"
	WorktreeStatusInUse   WorktreeStatus = "in-use"
	WorktreeStatusCorrupt WorktreeStatus = "corrupt"
)

type Worktree struct {
	ID        uuid.UUID      `db:"id"`
	RepoID    uuid.UUID      `db:"repo_id"`
	Name      string         `db:"name"`
	Path      string         `db:"path"`
	Status    WorktreeStatus `db:"status"`
	LeasedAt  *time.Time     `db:"leased_at"`
	Branch    *string        `db:"branch"`
	CreatedAt time.Time      `db:"created_at"`
}

func NewWorktree(repoID uuid.UUID, name, path string) *Worktree {
	return &Worktree{
		ID:        uuid.New(),
		RepoID:    repoID,
		Name:      name,
		Path:      path,
		Status:    WorktreeStatusIdle,
		LeasedAt:  nil,
		CreatedAt: time.Now(),
	}
}
