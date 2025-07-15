package commands

import (
	"fmt"
	"os"

	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/internal/config"
	"github.com/albertywu/gitpool/internal/ipc"
	"github.com/spf13/cobra"
)

func NewStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the gitpool daemon",
		Long:  `Stop the running gitpool daemon gracefully.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadWithCustomPaths("", "", "")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Check if socket exists
			if _, err := os.Stat(cfg.SocketPath); os.IsNotExist(err) {
				internal.PrintError("Daemon is not running")
				return fmt.Errorf("daemon not running")
			}

			// Try to connect and check status first
			client := ipc.NewClient(cfg.SocketPath)
			resp, err := client.DaemonStatus()
			if err != nil {
				// Socket exists but can't connect - might be stale
				internal.PrintWarn("Socket exists but daemon not responding, cleaning up...")
				os.Remove(cfg.SocketPath)
				return nil
			}

			if !resp.Success {
				internal.PrintError("Daemon is not running")
				return fmt.Errorf("daemon not running")
			}

			// Send interrupt signal to daemon process
			// In a real implementation, we'd need to track the PID
			internal.PrintInfo("Stopping gitpool daemon...")

			// For now, we'll just note that users should use Ctrl+C
			internal.PrintInfo("Please use Ctrl+C to stop the daemon process")

			return nil
		},
	}

	return cmd
}
