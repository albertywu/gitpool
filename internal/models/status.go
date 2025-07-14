package models

import (
	"time"

	"github.com/google/uuid"
)

type ReconcilerRun struct {
	ID      uuid.UUID `db:"id"`
	RunTime time.Time `db:"run_time"`
	Created int       `db:"created"`
	Cleaned int       `db:"cleaned"`
}

type DaemonStatus struct {
	Running        bool
	SocketPath     string
	LastReconciler *time.Time
	Repositories   int
}

type PoolStatus struct {
	RepoName  string
	Total     int
	InUse     int
	Idle      int
	Max       int
	LastFetch *time.Time
}
