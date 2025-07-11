package models

// WorktreeDetail contains a worktree with its associated repository information
type WorktreeDetail struct {
	Worktree   *Worktree
	Repository *Repository
}
