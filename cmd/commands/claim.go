package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uber/treefarm/config"
	"github.com/uber/treefarm/internal"
	"github.com/uber/treefarm/ipc"
)

var (
	claimRepo       string
	claimOutputPath bool
)

func NewClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim",
		Short: "Claim a worktree from the pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			if claimRepo == "" {
				return fmt.Errorf("--repo flag is required")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			req := ipc.ClaimRequest{
				RepoName:   claimRepo,
				OutputPath: claimOutputPath,
			}

			resp, err := client.Claim(req)
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to claim worktree: %s", resp.Error)
				return fmt.Errorf("claim failed")
			}

			// Print the result (path or name)
			fmt.Println(resp.Data)

			return nil
		},
	}

	cmd.Flags().StringVar(&claimRepo, "repo", "", "Repository name")
	cmd.Flags().BoolVar(&claimOutputPath, "output-path", false, "Output full path instead of worktree name")
	cmd.MarkFlagRequired("repo")

	return cmd
}
