package commands

import (
	"encoding/json"
	"fmt"

	"github.com/albertywu/gitpool/config"
	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/ipc"
	"github.com/albertywu/gitpool/models"
	"github.com/spf13/cobra"
)

var showFormat string

func NewShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <worktree-id>",
		Short: "Show details about a specific worktree",
		Long: `Display detailed information about a worktree including its path, status, branch, and claim time.
Use --format flag to get specific fields for scripting.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			worktreeID := args[0]

			cfg, err := config.LoadWithCustomPaths("", "", "")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			req := ipc.ShowRequest{
				WorktreeID: worktreeID,
			}

			resp, err := client.Show(req)
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to get worktree details: %s", resp.Error)
				return fmt.Errorf("show failed")
			}

			// Parse response
			data, _ := json.Marshal(resp.Data)
			var detail models.WorktreeDetail
			if err := json.Unmarshal(data, &detail); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			// Format output based on format flag
			switch showFormat {
			case "path":
				fmt.Println(detail.Worktree.Path)
			case "json":
				output := map[string]interface{}{
					"worktree_id": detail.Worktree.ID,
					"path":        detail.Worktree.Path,
					"repo":        detail.Repository.Name,
					"branch":      detail.Worktree.Branch,
					"status":      detail.Worktree.Status,
					"claimed_at":  detail.Worktree.LeasedAt,
				}
				jsonBytes, _ := json.MarshalIndent(output, "", "  ")
				fmt.Println(string(jsonBytes))
			default:
				// Default human-readable format
				fmt.Printf("Worktree ID: %s\n", detail.Worktree.ID)
				fmt.Printf("Path:        %s\n", detail.Worktree.Path)
				fmt.Printf("Repository:  %s\n", detail.Repository.Name)
				fmt.Printf("Status:      %s\n", detail.Worktree.Status)
				if detail.Worktree.Branch != nil {
					fmt.Printf("Branch:      %s\n", *detail.Worktree.Branch)
				}
				if detail.Worktree.LeasedAt != nil {
					fmt.Printf("Claimed at:  %s\n", detail.Worktree.LeasedAt.Format("2006-01-02 15:04:05"))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&showFormat, "format", "", "Output format: path, json (default: human-readable)")

	return cmd
}
