package commands

import (
	"fmt"

	"github.com/albertywu/gitpool/config"
	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/ipc"
	"github.com/spf13/cobra"
)

func NewRefreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh <repo-name>",
		Short: "Refresh a repository and update idle worktrees",
		Long: `Fetch latest changes from the remote repository and update all idle worktrees
to the latest commit on the default branch.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoName := args[0]

			cfg, err := config.LoadWithCustomPaths("", "", "")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			req := ipc.RefreshRequest{
				RepoName: repoName,
			}

			resp, err := client.Refresh(req)
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to refresh repository: %s", resp.Error)
				return fmt.Errorf("refresh failed")
			}

			internal.PrintInfo("Repository '%s' refreshed successfully", repoName)
			return nil
		},
	}
}
