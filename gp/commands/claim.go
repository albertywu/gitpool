package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/internal/config"
	"github.com/albertywu/gitpool/internal/ipc"
	"github.com/spf13/cobra"
)

func NewClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim <repo-name> <branch-name>",
		Short: "Claim a worktree from the pool",
		Long: `Claim an available worktree from the pool for the specified repository.

The branch name must be a valid git branch name and unique within the repository's worktrees.

The command outputs JSON with the worktree ID and path to STDOUT.
Error messages are printed to STDERR.

Example:
  gp claim my-app feature-xyz
  
Output:
  {
    "worktree_id": "a91b6fc1-4322-4b2f-8c1a-123456789abc",
    "path": "/home/user/.gitpool/worktrees/my-app/a91b6fc1-4322-4b2f-8c1a-123456789abc"
  }
  
Usage with jq:
  # Get just the path
  gp claim my-app feature-xyz | jq -r .path
  
  # CD into the worktree
  cd $(gp claim my-app feature-xyz | jq -r .path)`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoName := args[0]
			branch := args[1]

			// Validate branch name
			if err := validateBranchName(branch); err != nil {
				return fmt.Errorf("invalid branch name: %w", err)
			}

			cfg, err := config.LoadWithCustomPaths("", "", "")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			req := ipc.ClaimRequest{
				RepoName: repoName,
				Branch:   branch,
			}

			resp, err := client.Claim(req)
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to claim worktree: %s", resp.Error)
				return fmt.Errorf("claim failed")
			}

			// Parse the response to get both worktree ID and path
			// First, convert the response data to JSON
			jsonData, err := json.Marshal(resp.Data)
			if err != nil {
				return fmt.Errorf("failed to marshal response: %w", err)
			}

			var claimResp ipc.ClaimResponse
			if err := json.Unmarshal(jsonData, &claimResp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Create JSON output with worktree ID and path
			output := map[string]string{
				"worktree_id": claimResp.WorktreeID,
				"path":        claimResp.Path,
			}

			// Marshal and print JSON to STDOUT
			jsonOutput, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to create JSON output: %w", err)
			}
			fmt.Println(string(jsonOutput))

			return nil
		},
	}

	return cmd
}

// validateBranchName checks if a branch name is valid according to git rules
func validateBranchName(branch string) error {
	// Basic git branch name validation rules
	if branch == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Check for invalid characters
	invalidChars := []string{" ", "..", "~", "^", ":", "?", "*", "[", "\\", "@{"}
	for _, char := range invalidChars {
		if strings.Contains(branch, char) {
			return fmt.Errorf("branch name contains invalid character: %s", char)
		}
	}

	// Check for invalid patterns
	if branch[0] == '.' || branch[0] == '-' {
		return fmt.Errorf("branch name cannot start with '.' or '-'")
	}

	if branch[len(branch)-1] == '.' || branch[len(branch)-1] == '/' {
		return fmt.Errorf("branch name cannot end with '.' or '/'")
	}

	if branch == "@" {
		return fmt.Errorf("branch name cannot be '@'")
	}

	if strings.Contains(branch, "//") {
		return fmt.Errorf("branch name cannot contain consecutive slashes")
	}

	return nil
}
