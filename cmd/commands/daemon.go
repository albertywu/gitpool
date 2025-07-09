package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/uber/treefarm/config"
	"github.com/uber/treefarm/daemon"
	"github.com/uber/treefarm/internal"
	"github.com/uber/treefarm/ipc"
)

var (
	daemonFetchInterval string
)

func NewDaemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the treefarm daemon",
	}

	cmd.AddCommand(newDaemonStartCmd())
	cmd.AddCommand(newDaemonStatusCmd())

	return cmd
}

func newDaemonStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the treefarm daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Override with flags if provided
			if daemonFetchInterval != "" {
				interval, err := time.ParseDuration(daemonFetchInterval)
				if err != nil {
					return fmt.Errorf("invalid fetch interval: %w", err)
				}
				cfg.FetchInterval = interval
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

			return d.Start()
		},
	}

	cmd.Flags().StringVar(&daemonFetchInterval, "fetch-interval", "", "Global fetch interval (e.g., 15m)")

	return cmd
}

func newDaemonStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check daemon status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
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

			// Parse response data
			var status map[string]interface{}
			if err := json.Unmarshal(resp.Data.(json.RawMessage), &status); err != nil {
				return fmt.Errorf("failed to parse status: %w", err)
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
}
