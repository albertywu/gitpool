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
