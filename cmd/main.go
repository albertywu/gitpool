package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uber/treefarm/cmd/commands"
	"github.com/uber/treefarm/internal"
)

func main() {
	internal.InitLogger()

	rootCmd := &cobra.Command{
		Use:   "treefarm",
		Short: "Manage a pool of pre-initialized Git worktrees",
		Long: `treefarm is a CLI + daemon tool for managing a pool of pre-initialized Git worktrees.
It enables fast, disposable checkouts for builds, tests, and CI pipelines without repeated Git fetches.
Developers can instantly "claim" worktrees and "release" them back for reuse.`,
	}

	// Add subcommands
	rootCmd.AddCommand(commands.NewDaemonCmd())
	rootCmd.AddCommand(commands.NewRepoCmd())
	rootCmd.AddCommand(commands.NewClaimCmd())
	rootCmd.AddCommand(commands.NewReleaseCmd())
	rootCmd.AddCommand(commands.NewPoolCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
