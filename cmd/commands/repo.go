package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uber/treefarm/config"
	"github.com/uber/treefarm/internal"
	"github.com/uber/treefarm/ipc"
	"github.com/uber/treefarm/models"
)

var (
	repoMaxWorktrees  int
	repoDefaultBranch string
	repoFetchInterval string
)

func NewRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repositories",
	}

	cmd.AddCommand(newRepoAddCmd())
	cmd.AddCommand(newRepoListCmd())
	cmd.AddCommand(newRepoRemoveCmd())

	return cmd
}

func newRepoAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <path>",
		Short: "Register a new repository",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			path := args[1]

			// Parse fetch interval
			fetchIntervalMinutes := 10
			if repoFetchInterval != "" {
				var minutes int
				if _, err := fmt.Sscanf(repoFetchInterval, "%dm", &minutes); err != nil {
					return fmt.Errorf("invalid fetch interval format (use format like '5m')")
				}
				fetchIntervalMinutes = minutes
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			req := ipc.RepoAddRequest{
				Name:          name,
				Path:          path,
				MaxWorktrees:  repoMaxWorktrees,
				DefaultBranch: repoDefaultBranch,
				FetchInterval: fetchIntervalMinutes,
			}

			resp, err := client.RepoAdd(req)
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to add repository: %s", resp.Error)
				return fmt.Errorf("add repository failed")
			}

			internal.PrintInfo("Repository added successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&repoMaxWorktrees, "max", 8, "Maximum number of worktrees")
	cmd.Flags().StringVar(&repoDefaultBranch, "default-branch", "main", "Default branch to checkout")
	cmd.Flags().StringVar(&repoFetchInterval, "fetch-interval", "10m", "Fetch interval (e.g., 5m, 10m)")

	return cmd
}

func newRepoListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			resp, err := client.RepoList()
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to list repositories: %s", resp.Error)
				return fmt.Errorf("list repositories failed")
			}

			// Parse response
			data, _ := json.Marshal(resp.Data)
			var repos []*models.Repository
			if err := json.Unmarshal(data, &repos); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			if len(repos) == 0 {
				fmt.Println("No repositories registered")
				return nil
			}

			// Print table
			w := internal.NewTabWriter()
			fmt.Fprintln(w, "NAME\tPATH\tMAX\tDEFAULT BRANCH\tFETCH INTERVAL")
			for _, repo := range repos {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%dm\n",
					repo.Name, repo.Path, repo.MaxWorktrees,
					repo.DefaultBranch, repo.FetchInterval)
			}
			w.Flush()

			return nil
		},
	}
}

func newRepoRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := ipc.NewClient(cfg.SocketPath)
			resp, err := client.RepoRemove(name)
			if err != nil {
				return fmt.Errorf("failed to communicate with daemon: %w", err)
			}

			if !resp.Success {
				internal.PrintError("Failed to remove repository: %s", resp.Error)
				return fmt.Errorf("remove repository failed")
			}

			internal.PrintInfo("Repository removed successfully")
			return nil
		},
	}
}
