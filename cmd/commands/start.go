package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/albertywu/gitpool/config"
	"github.com/albertywu/gitpool/daemon"
	"github.com/albertywu/gitpool/internal"
)

var (
	startConfigDir   string
	startWorktreeDir string
	startSocketPath  string
)

func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the gitpool daemon",
		Long:  `Start the gitpool daemon to manage the worktree pool in the background.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadWithCustomPaths(startConfigDir, startWorktreeDir, startSocketPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Check if daemon is already running
			if daemon.CheckDaemonRunning(cfg.SocketPath) {
				internal.PrintError("Another instance is already running (socket lock exists)")
				return fmt.Errorf("daemon already running")
			}

			// Create and start daemon
			d, err := daemon.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to create daemon: %w", err)
			}

			internal.PrintInfo("Starting gitpool daemon...")
			return d.Start()
		},
	}

	cmd.Flags().StringVar(&startConfigDir, "config-dir", "", "Custom config directory")
	cmd.Flags().StringVar(&startWorktreeDir, "worktree-dir", "", "Custom worktree directory")
	cmd.Flags().StringVar(&startSocketPath, "socket-path", "", "Custom socket path")

	return cmd
}