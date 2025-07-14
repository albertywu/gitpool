package commands

import (
	"fmt"

	"github.com/albertywu/gitpool/internal/config"
	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/internal/ipc"
	"github.com/spf13/cobra"
)

var (
	trackMaxWorktrees  int
	trackDefaultBranch string
)

func NewTrackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "track <repo-name> <repo-path>",
		Short: "Track a new repository",
		Long:  `Track a new Git repository with gitpool to create and manage a pool of worktrees.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			path := args[1]

			cfg, err := config.LoadWithCustomPaths("", "", "")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			req := ipc.RepoAddRequest{
				Name:          name,
				Path:          path,
				MaxWorktrees:  trackMaxWorktrees,
				DefaultBranch: trackDefaultBranch,
			}

			resp, err := client.RepoAdd(req)
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to track repository: %s", resp.Error)
				return fmt.Errorf("track repository failed")
			}

			internal.PrintInfo("Repository '%s' tracked successfully", name)
			return nil
		},
	}

	cmd.Flags().IntVar(&trackMaxWorktrees, "max", 8, "Maximum number of worktrees")
	cmd.Flags().StringVar(&trackDefaultBranch, "default-branch", "main", "Default branch to checkout")

	return cmd
}
