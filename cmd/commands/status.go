package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/albertywu/gitpool/config"
	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/ipc"
	"github.com/albertywu/gitpool/models"
)

func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show pool usage statistics",
		Long:  `Display the current status of all worktree pools, including usage statistics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadWithCustomPaths("", "", "")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			req := ipc.PoolStatusRequest{}

			resp, err := client.PoolStatus(req)
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to get pool status: %s", resp.Error)
				return fmt.Errorf("pool status failed")
			}

			// Parse response
			data, _ := json.Marshal(resp.Data)
			var statuses []*models.PoolStatus
			if err := json.Unmarshal(data, &statuses); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			if len(statuses) == 0 {
				fmt.Println("No repositories in pool")
				return nil
			}

			// Print table
			w := internal.NewTabWriter()
			fmt.Fprintln(w, "REPO\tTOTAL\tIN-USE\tIDLE\tMAX\tLAST FETCH")
			for _, status := range statuses {
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%s\n",
					status.RepoName, status.Total, status.InUse,
					status.Idle, status.Max, internal.FormatTime(status.LastFetch))
			}
			w.Flush()

			return nil
		},
	}

	return cmd
}