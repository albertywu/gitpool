package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uber/treefarm/config"
	"github.com/uber/treefarm/internal"
	"github.com/uber/treefarm/ipc"
)

func NewReleaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "release <worktree-id>",
		Short: "Release a worktree back to the pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			worktreeID := args[0]

			cfg, err := config.LoadWithCustomPaths("", "", "")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			req := ipc.ReleaseRequest{
				WorktreeID: worktreeID,
			}

			resp, err := client.Release(req)
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to release worktree: %s", resp.Error)
				return fmt.Errorf("release failed")
			}

			internal.PrintInfo("Worktree released successfully")
			return nil
		},
	}
}
