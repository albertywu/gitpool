package commands

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/albertywu/gitpool/internal"
	"github.com/albertywu/gitpool/internal/config"
	"github.com/albertywu/gitpool/internal/ipc"
	"github.com/albertywu/gitpool/internal/models"
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

			// Sort worktrees: first by status (claimed before unclaimed), then by repo name, then by creation time
			sort.Slice(details, func(i, j int) bool {
				// First, prioritize claimed (IN-USE) worktrees
				iClaimed := details[i].Worktree.Status == models.WorktreeStatusInUse
				jClaimed := details[j].Worktree.Status == models.WorktreeStatusInUse

				if iClaimed != jClaimed {
					return iClaimed // true (claimed) comes before false (unclaimed)
				}

				// Then, compare by repository name
				if details[i].Repository.Name != details[j].Repository.Name {
					return details[i].Repository.Name < details[j].Repository.Name
				}

				// If same repo and same status, sort by creation time (newer first)
				return details[i].Worktree.CreatedAt.After(details[j].Worktree.CreatedAt)
			})

			// Calculate column widths based on data
			// Start with header lengths as minimum
			idWidth := len("ID")
			worktreeWidth := len("WORKTREE")
			repoWidth := len("REPO")
			claimedAtWidth := len("CLAIMED_AT")

			// Find maximum widths based on actual data
			for _, detail := range details {
				wt := detail.Worktree
				repo := detail.Repository

				// ID column
				if len(wt.Name) > idWidth {
					idWidth = len(wt.Name)
				}

				// Worktree column
				worktreeLen := 0
				if wt.Status == models.WorktreeStatusInUse && wt.Branch != nil && *wt.Branch != "" {
					worktreeLen = len(*wt.Branch)
				} else {
					worktreeLen = len("UNCLAIMED")
				}
				if worktreeLen > worktreeWidth {
					worktreeWidth = worktreeLen
				}

				// Repo column
				if len(repo.Name) > repoWidth {
					repoWidth = len(repo.Name)
				}

				// Claimed at column
				var claimedAtLen int
				if wt.Status == models.WorktreeStatusInUse && wt.LeasedAt != nil {
					timeSince := time.Since(*wt.LeasedAt)
					if timeSince < time.Minute {
						claimedAtLen = len("just now")
					} else if timeSince < time.Hour {
						claimedAtLen = len(fmt.Sprintf("%dm ago", int(timeSince.Minutes())))
					} else if timeSince < 24*time.Hour {
						claimedAtLen = len(fmt.Sprintf("%dh ago", int(timeSince.Hours())))
					} else {
						claimedAtLen = len(fmt.Sprintf("%dd ago", int(timeSince.Hours()/24)))
					}
				} else {
					claimedAtLen = len("-")
				}
				if claimedAtLen > claimedAtWidth {
					claimedAtWidth = claimedAtLen
				}
			}

			// Add some padding
			idWidth += 2
			worktreeWidth += 2
			repoWidth += 2
			claimedAtWidth += 2

			// Print beautiful header
			fmt.Printf("\n%s%sWorktree Pool Status%s\n", colorBold, colorCyan, colorReset)
			totalWidth := idWidth + worktreeWidth + repoWidth + claimedAtWidth + 6 // 6 for spacing
			fmt.Printf("%s%s%s\n\n", colorGray, strings.Repeat("─", totalWidth), colorReset)

			// Helper function to pad string to fixed width
			padRight := func(s string, width int) string {
				if len(s) >= width {
					return s[:width]
				}
				return s + strings.Repeat(" ", width-len(s))
			}

			// Print table header
			fmt.Printf("%s%s%-*s  %-*s  %-*s  %-*s%s\n",
				colorBold, colorGray,
				idWidth, "ID",
				worktreeWidth, "WORKTREE",
				repoWidth, "REPO",
				claimedAtWidth, "CLAIMED_AT",
				colorReset)

			// Print separator
			fmt.Printf("%s%s  %s  %s  %s%s\n",
				colorGray,
				strings.Repeat("─", idWidth),
				strings.Repeat("─", worktreeWidth),
				strings.Repeat("─", repoWidth),
				strings.Repeat("─", claimedAtWidth),
				colorReset)

			// Print worktrees
			for _, detail := range details {
				wt := detail.Worktree
				repo := detail.Repository

				// Calculate time since claimed
				var claimedAtDisplay string
				if wt.Status == models.WorktreeStatusInUse && wt.LeasedAt != nil {
					timeSince := time.Since(*wt.LeasedAt)
					if timeSince < time.Minute {
						claimedAtDisplay = "just now"
					} else if timeSince < time.Hour {
						claimedAtDisplay = fmt.Sprintf("%dm ago", int(timeSince.Minutes()))
					} else if timeSince < 24*time.Hour {
						claimedAtDisplay = fmt.Sprintf("%dh ago", int(timeSince.Hours()))
					} else {
						claimedAtDisplay = fmt.Sprintf("%dd ago", int(timeSince.Hours()/24))
					}
				} else {
					// Show dash for unclaimed worktrees
					claimedAtDisplay = "-"
				}

				// Format worktree display based on status
				var worktreeDisplay string
				var worktreeColor string

				if wt.Status == models.WorktreeStatusInUse && wt.Branch != nil && *wt.Branch != "" {
					// Show branch name in yellow for claimed worktrees
					worktreeDisplay = *wt.Branch
					worktreeColor = colorYellow
				} else {
					// Show "UNCLAIMED" in gray for idle worktrees
					worktreeDisplay = "UNCLAIMED"
					worktreeColor = colorGray
				}

				// Add terminal hyperlink to worktree path
				// Format: OSC 8 ; params ; URI ST display_text OSC 8 ; ; ST
				// OSC = \033]  ST = \033\\
				terminalLink := fmt.Sprintf("\033]8;;file://%s\033\\%s%s%s\033]8;;\033\\",
					wt.Path, worktreeColor, padRight(worktreeDisplay, worktreeWidth), colorReset)

				// Format the row with fixed widths
				fmt.Printf("%s%-*s%s  %s  %s%-*s%s  %s%-*s%s\n",
					colorBlue, idWidth, wt.Name, colorReset,
					terminalLink,
					colorPurple, repoWidth, repo.Name, colorReset,
					colorGray, claimedAtWidth, claimedAtDisplay, colorReset)
			}

			// Print summary
			fmt.Printf("\n%s%s%s\n", colorGray, strings.Repeat("─", totalWidth), colorReset)
			fmt.Printf("%sSummary:%s Total: %s%d%s worktrees\n\n",
				colorBold, colorReset,
				colorBold, len(details), colorReset)

			return nil
		},
	}
}
