package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/albertywu/gitpool/config"
	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/ipc"
	"github.com/albertywu/gitpool/models"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all worktrees with their details",
		Long:  `Display all worktrees in the pool with their repository information and status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadWithCustomPaths("", "", "")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			resp, err := client.WorktreeList()
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to list worktrees: %s", resp.Error)
				return fmt.Errorf("list worktrees failed")
			}

			// Parse response
			data, _ := json.Marshal(resp.Data)
			var details []*models.WorktreeDetail
			if err := json.Unmarshal(data, &details); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			if len(details) == 0 {
				fmt.Println("No worktrees in pool")
				return nil
			}

			// Print beautiful header
			fmt.Printf("\n%s%sWorktree Pool Status%s\n", colorBold, colorCyan, colorReset)
			fmt.Printf("%s%s%s\n\n", colorGray, strings.Repeat("─", 80), colorReset)

			// Print table header
			w := internal.NewTabWriter()
			fmt.Fprintf(w, "%s%sID\tWORKSPACE\tREPO\tSTATUS\tMAX\tBRANCH\tLAST FETCH%s\n", 
				colorBold, colorGray, colorReset)
			
			// Print separator
			fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\t%s\t%s\t%s%s\n",
				colorGray,
				strings.Repeat("─", 20),
				strings.Repeat("─", 40),
				strings.Repeat("─", 15),
				strings.Repeat("─", 10),
				strings.Repeat("─", 5),
				strings.Repeat("─", 10),
				strings.Repeat("─", 10),
				colorReset)

			// Print worktrees
			for _, detail := range details {
				wt := detail.Worktree
				repo := detail.Repository
				
				// Choose color based on status
				statusColor := colorGreen
				statusText := "IDLE"
				if wt.Status == models.WorktreeStatusInUse {
					statusColor = colorYellow
					statusText = "IN-USE"
				} else if wt.Status == models.WorktreeStatusCorrupt {
					statusColor = colorRed
					statusText = "CORRUPT"
				}

				// Calculate time since last fetch
				var lastFetchDisplay string
				if repo.LastFetchTime != nil {
					timeSince := time.Since(*repo.LastFetchTime)
					if timeSince < time.Minute {
						lastFetchDisplay = "just now"
					} else if timeSince < time.Hour {
						lastFetchDisplay = fmt.Sprintf("%dm ago", int(timeSince.Minutes()))
					} else if timeSince < 24*time.Hour {
						lastFetchDisplay = fmt.Sprintf("%dh ago", int(timeSince.Hours()))
					} else {
						lastFetchDisplay = fmt.Sprintf("%dd ago", int(timeSince.Hours()/24))
					}
				} else {
					lastFetchDisplay = "never"
				}
				
				// Format workspace display based on status
				var workspaceDisplay string
				var workspaceColor string
				
				if wt.Status == models.WorktreeStatusInUse && wt.Branch != nil && *wt.Branch != "" {
					// Show branch name in yellow for claimed workspaces
					workspaceDisplay = *wt.Branch
					workspaceColor = colorYellow
				} else {
					// Show "UNCLAIMED" in gray for idle workspaces
					workspaceDisplay = "UNCLAIMED"
					workspaceColor = colorGray
				}
				
				// Add terminal hyperlink to workspace path
				// Format: OSC 8 ; params ; URI ST display_text OSC 8 ; ; ST
				// OSC = \033]  ST = \033\\
				terminalLink := fmt.Sprintf("\033]8;;file://%s\033\\%s%s%s\033]8;;\033\\", 
					wt.Path, workspaceColor, workspaceDisplay, colorReset)
				
				// Format the row
				fmt.Fprintf(w, "%s%s%s\t%s\t%s%s%s\t%s%s%s\t%d\t%s%s%s\t%s%s%s\n",
					colorBlue, wt.Name, colorReset,
					terminalLink,
					colorPurple, repo.Name, colorReset,
					statusColor, statusText, colorReset,
					repo.MaxWorktrees,
					colorCyan, repo.DefaultBranch, colorReset,
					colorGray, lastFetchDisplay, colorReset)
			}
			
			w.Flush()
			
			// Print summary
			idle := 0
			inUse := 0
			corrupt := 0
			for _, detail := range details {
				switch detail.Worktree.Status {
				case models.WorktreeStatusIdle:
					idle++
				case models.WorktreeStatusInUse:
					inUse++
				case models.WorktreeStatusCorrupt:
					corrupt++
				}
			}
			
			fmt.Printf("\n%s%s%s\n", colorGray, strings.Repeat("─", 80), colorReset)
			fmt.Printf("%sSummary:%s Total: %s%d%s | Idle: %s%d%s | In-Use: %s%d%s | Corrupt: %s%d%s\n\n",
				colorBold, colorReset,
				colorBold, len(details), colorReset,
				colorGreen, idle, colorReset,
				colorYellow, inUse, colorReset,
				colorRed, corrupt, colorReset)

			return nil
		},
	}
}