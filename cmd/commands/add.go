package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/albertywu/gitpool/config"
	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/ipc"
)

var (
	addMaxWorktrees  int
	addDefaultBranch string
)

func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <repo-name> <repo-path>",
		Short: "Register a new repository",
		Long:  `Register a new Git repository with gitpool to create and manage a pool of worktrees.`,
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
				MaxWorktrees:  addMaxWorktrees,
				DefaultBranch: addDefaultBranch,
				FetchInterval: 60, // Default value, will be ignored by daemon
			}

			resp, err := client.RepoAdd(req)
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to add repository: %s", resp.Error)
				return fmt.Errorf("add repository failed")
			}

			internal.PrintInfo("Repository '%s' added successfully", name)
			return nil
		},
	}

	cmd.Flags().IntVar(&addMaxWorktrees, "max", 8, "Maximum number of worktrees")
	cmd.Flags().StringVar(&addDefaultBranch, "default-branch", "main", "Default branch to checkout")

	return cmd
}