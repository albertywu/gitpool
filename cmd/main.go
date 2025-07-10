package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/albertywu/gitpool/cmd/commands"
	"github.com/albertywu/gitpool/internal"
)

func main() {
	internal.InitLogger()

	rootCmd := &cobra.Command{
		Use:   "gitpool",
		Short: "Manage a pool of pre-initialized Git worktrees",
		Long: `gitpool is a CLI + daemon tool for managing a pool of pre-initialized Git worktrees.
It enables fast, disposable checkouts for builds, tests, and CI pipelines without repeated Git fetches.
Developers can instantly "claim" worktrees and "release" them back for reuse.`,
	}

	// Add simplified top-level commands
	rootCmd.AddCommand(commands.NewStartCmd())
	rootCmd.AddCommand(commands.NewStopCmd())
	rootCmd.AddCommand(commands.NewStatusCmd())
	rootCmd.AddCommand(commands.NewAddCmd())
	rootCmd.AddCommand(commands.NewRemoveCmd())
	rootCmd.AddCommand(commands.NewClaimCmd())
	rootCmd.AddCommand(commands.NewReleaseCmd())
	
	// Keep list command for repositories
	rootCmd.AddCommand(commands.NewListCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
