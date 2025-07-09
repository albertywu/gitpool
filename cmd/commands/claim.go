package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/albertywu/gitpool/config"
	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/ipc"
)

var (
	claimRepo string
)

func NewClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim",
		Short: "Claim a worktree from the pool",
		Long: `Claim an available worktree from the pool for the specified repository.

Output format (two lines):
  <worktree-id>
  <absolute-path>

Example:
  my-app-a91b6fc1
  /home/user/.gitpool/worktrees/my-app-a91b6fc1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if claimRepo == "" {
				return fmt.Errorf("--repo flag is required")
			}

			cfg, err := config.LoadWithCustomPaths("", "", "")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			req := ipc.ClaimRequest{
				RepoName: claimRepo,
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
			var claimResp ipc.ClaimResponse
			if err := json.Unmarshal(resp.Data.(json.RawMessage), &claimResp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Print worktree ID and path on separate lines
			fmt.Printf("%s\n%s\n", claimResp.WorktreeID, claimResp.Path)

			return nil
		},
	}

	cmd.Flags().StringVar(&claimRepo, "repo", "", "Repository name")
	cmd.MarkFlagRequired("repo")

	return cmd
}
