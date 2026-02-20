package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spk",
	Short: "spk — workspace CLI for multi-repo development",
	Long: `spk is a workspace-oriented CLI that keeps multiple repositories
in sync, manages AWS credentials at the workspace level, and provides
dependency-aware builds across projects.

Get started:
  spk create workspace ./my-project
  cd my-project
  spk use org/repo-name
  spk sync`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
