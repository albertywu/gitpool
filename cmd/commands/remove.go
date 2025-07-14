package commands

import (
	"fmt"

	"github.com/albertywu/gitpool/internal/config"
	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/internal/ipc"
	"github.com/spf13/cobra"
)

func NewRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <repo-name>",
		Short: "Remove a repository",
		Long:  `Remove a repository from gitpool and clean up all its worktrees.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.LoadWithCustomPaths("", "", "")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			resp, err := client.RepoRemove(name)
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to remove repository: %s", resp.Error)
				return fmt.Errorf("remove repository failed")
			}

			internal.PrintInfo("Repository '%s' removed successfully", name)
			return nil
		},
	}
}
