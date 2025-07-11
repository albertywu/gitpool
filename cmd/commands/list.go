package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/albertywu/gitpool/config"
	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/ipc"
	"github.com/albertywu/gitpool/models"
	"github.com/spf13/cobra"
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

			// Calculate column widths based on data
			// Start with header lengths as minimum
			idWidth := len("ID")
			workspaceWidth := len("WORKSPACE")
			repoWidth := len("REPO")
			statusWidth := len("STATUS")
			maxWidth := len("MAX")
			branchWidth := len("BRANCH")
			lastSyncWidth := len("LAST_SYNC")

			// Find maximum widths based on actual data
			for _, detail := range details {
				wt := detail.Worktree
				repo := detail.Repository

				// ID column
				if len(wt.Name) > idWidth {
					idWidth = len(wt.Name)
				}

				// Workspace column
				workspaceLen := 0
				if wt.Status == models.WorktreeStatusInUse && wt.Branch != nil && *wt.Branch != "" {
					workspaceLen = len(*wt.Branch)
				} else {
					workspaceLen = len("UNCLAIMED")
				}
				if workspaceLen > workspaceWidth {
					workspaceWidth = workspaceLen
				}

				// Repo column
				if len(repo.Name) > repoWidth {
					repoWidth = len(repo.Name)
				}

				// Status column
				statusLen := 0
				switch wt.Status {
				case models.WorktreeStatusInUse:
					statusLen = len("IN-USE")
				case models.WorktreeStatusCorrupt:
					statusLen = len("CORRUPT")
				default:
					statusLen = len("IDLE")
				}
				if statusLen > statusWidth {
					statusWidth = statusLen
				}

				// MAX column
				maxStr := fmt.Sprintf("%d", repo.MaxWorktrees)
				if len(maxStr) > maxWidth {
					maxWidth = len(maxStr)
				}

				// Branch column
				if len(repo.DefaultBranch) > branchWidth {
					branchWidth = len(repo.DefaultBranch)
				}

				// Last sync column
				var lastFetchLen int
				if repo.LastFetchTime != nil {
					timeSince := time.Since(*repo.LastFetchTime)
					if timeSince < time.Minute {
						lastFetchLen = len("just now")
					} else if timeSince < time.Hour {
						lastFetchLen = len(fmt.Sprintf("%dm ago", int(timeSince.Minutes())))
					} else if timeSince < 24*time.Hour {
						lastFetchLen = len(fmt.Sprintf("%dh ago", int(timeSince.Hours())))
					} else {
						lastFetchLen = len(fmt.Sprintf("%dd ago", int(timeSince.Hours()/24)))
					}
				} else {
					lastFetchLen = len("never")
				}
				if lastFetchLen > lastSyncWidth {
					lastSyncWidth = lastFetchLen
				}
			}

			// Add some padding
			idWidth += 2
			workspaceWidth += 2
			repoWidth += 2
			statusWidth += 2
			maxWidth += 2
			branchWidth += 2
			lastSyncWidth += 2

			// Print beautiful header
			fmt.Printf("\n%s%sWorktree Pool Status%s\n", colorBold, colorCyan, colorReset)
			totalWidth := idWidth + workspaceWidth + repoWidth + statusWidth + maxWidth + branchWidth + lastSyncWidth + 12 // 12 for spacing
			fmt.Printf("%s%s%s\n\n", colorGray, strings.Repeat("─", totalWidth), colorReset)

			// Helper function to pad string to fixed width
			padRight := func(s string, width int) string {
				if len(s) >= width {
					return s[:width]
				}
				return s + strings.Repeat(" ", width-len(s))
			}

			// Print table header
			fmt.Printf("%s%s%-*s  %-*s  %-*s  %-*s  %-*s  %-*s  %-*s%s\n",
				colorBold, colorGray,
				idWidth, "ID",
				workspaceWidth, "WORKSPACE",
				repoWidth, "REPO",
				statusWidth, "STATUS",
				maxWidth, "MAX",
				branchWidth, "BRANCH",
				lastSyncWidth, "LAST_SYNC",
				colorReset)

			// Print separator
			fmt.Printf("%s%s  %s  %s  %s  %s  %s  %s%s\n",
				colorGray,
				strings.Repeat("─", idWidth),
				strings.Repeat("─", workspaceWidth),
				strings.Repeat("─", repoWidth),
				strings.Repeat("─", statusWidth),
				strings.Repeat("─", maxWidth),
				strings.Repeat("─", branchWidth),
				strings.Repeat("─", lastSyncWidth),
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
					wt.Path, workspaceColor, padRight(workspaceDisplay, workspaceWidth), colorReset)

				// Format the row with fixed widths
				fmt.Printf("%s%-*s%s  %s  %s%-*s%s  %s%-*s%s  %-*d  %s%-*s%s  %s%-*s%s\n",
					colorBlue, idWidth, wt.Name, colorReset,
					terminalLink,
					colorPurple, repoWidth, repo.Name, colorReset,
					statusColor, statusWidth, statusText, colorReset,
					maxWidth, repo.MaxWorktrees,
					colorCyan, branchWidth, repo.DefaultBranch, colorReset,
					colorGray, lastSyncWidth, lastFetchDisplay, colorReset)
			}

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

			fmt.Printf("\n%s%s%s\n", colorGray, strings.Repeat("─", totalWidth), colorReset)
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
