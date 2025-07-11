package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/albertywu/gitpool/config"
	"github.com/albertywu/gitpool/daemon"
	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/ipc"
	"github.com/spf13/cobra"
)

func NewDaemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the gitpool daemon",
	}

	cmd.AddCommand(newDaemonStartCmd())
	cmd.AddCommand(newDaemonStatusCmd())
	cmd.AddCommand(newDaemonStopCmd())

	return cmd
}

var (
	daemonConfigDir   string
	daemonWorktreeDir string
	daemonSocketPath  string
	daemonBackground  bool
	daemonPidFile     string
)

func newDaemonStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the gitpool daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadWithCustomPaths(daemonConfigDir, daemonWorktreeDir, daemonSocketPath)
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

			// TODO: Implement background mode
			if daemonBackground {
				internal.PrintError("Background mode not yet implemented")
				return fmt.Errorf("background mode not supported")
			}

			return d.Start()
		},
	}

	cmd.Flags().StringVar(&daemonConfigDir, "config-dir", "", "Custom config directory")
	cmd.Flags().StringVar(&daemonWorktreeDir, "worktree-dir", "", "Custom worktree directory")
	cmd.Flags().StringVar(&daemonSocketPath, "socket-path", "", "Custom socket path")
	cmd.Flags().BoolVar(&daemonBackground, "background", false, "Run daemon in background")
	cmd.Flags().StringVar(&daemonPidFile, "pid-file", "", "PID file for background daemon")

	return cmd
}

func newDaemonStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check daemon status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadWithCustomPaths(daemonConfigDir, daemonWorktreeDir, daemonSocketPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			resp, err := client.DaemonStatus()
			if err != nil {
				internal.PrintError("Daemon is not running")
				return err
			}

			if !resp.Success {
				internal.PrintError("Failed to get daemon status: %s", resp.Error)
				return fmt.Errorf("daemon status failed")
			}

			// Parse response data - resp.Data is already the correct type
			var status map[string]interface{}
			if rawMsg, ok := resp.Data.(json.RawMessage); ok {
				if err := json.Unmarshal(rawMsg, &status); err != nil {
					return fmt.Errorf("failed to parse status: %w", err)
				}
			} else if statusMap, ok := resp.Data.(map[string]interface{}); ok {
				status = statusMap
			} else {
				return fmt.Errorf("unexpected response data type: %T", resp.Data)
			}

			internal.PrintInfo("Daemon is running")
			internal.PrintInfo("Listening on %s", status["socket_path"])

			if lastRun, ok := status["last_reconciler"].(string); ok && lastRun != "" {
				t, _ := time.Parse(time.RFC3339, lastRun)
				internal.PrintInfo("Last reconciler run: %s", internal.FormatTime(&t))
			}

			internal.PrintInfo("Managed repositories: %v", status["repositories"])

			if uptime, ok := status["uptime"].(string); ok {
				internal.PrintInfo("Uptime: %s", uptime)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&daemonConfigDir, "config-dir", "", "Custom config directory")
	cmd.Flags().StringVar(&daemonWorktreeDir, "worktree-dir", "", "Custom worktree directory")
	cmd.Flags().StringVar(&daemonSocketPath, "socket-path", "", "Custom socket path")

	return cmd
}

func newDaemonStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the gitpool daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadWithCustomPaths(daemonConfigDir, daemonWorktreeDir, daemonSocketPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if !daemon.CheckDaemonRunning(cfg.SocketPath) {
				internal.PrintError("Daemon is not running")
				return fmt.Errorf("daemon not running")
			}

			// TODO: Implement proper stop via IPC
			internal.PrintError("Daemon stop not yet implemented")
			internal.PrintInfo("Use 'kill' command to stop daemon manually")
			return fmt.Errorf("daemon stop not supported")
		},
	}

	cmd.Flags().StringVar(&daemonConfigDir, "config-dir", "", "Custom config directory")
	cmd.Flags().StringVar(&daemonWorktreeDir, "worktree-dir", "", "Custom worktree directory")
	cmd.Flags().StringVar(&daemonSocketPath, "socket-path", "", "Custom socket path")

	return cmd
}
