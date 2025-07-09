package models

import (
	"testing"
	"time"
)

func TestNewRepository(t *testing.T) {
	repo := NewRepository("test-repo", "/path/to/repo", "main", 8, 10)

	if repo.Name != "test-repo" {
		t.Errorf("expected name 'test-repo', got %s", repo.Name)
	}

	if repo.Path != "/path/to/repo" {
		t.Errorf("expected path '/path/to/repo', got %s", repo.Path)
	}

	if repo.DefaultBranch != "main" {
		t.Errorf("expected default branch 'main', got %s", repo.DefaultBranch)
	}

	if repo.MaxWorktrees != 8 {
		t.Errorf("expected max worktrees 8, got %d", repo.MaxWorktrees)
	}

	if repo.FetchInterval != 10 {
		t.Errorf("expected fetch interval 10, got %d", repo.FetchInterval)
	}

	if time.Since(repo.CreatedAt) > 1*time.Second {
		t.Errorf("created_at should be recent")
	}
}
