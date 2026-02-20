package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "spark-cli",
	Short:   "spark-cli — multi-repo workspace CLI",
	Version: Version,
	Long: `spark-cli manages multi-repo workspaces with shared environment and smart builds.

Core Commands:
  create workspace <path>   Create a new workspace
  use <repo>                Add a repo to the workspace
  sync                      Sync repos + refresh .env (auto-login)
  run <script>              Run npm/gradle script (build, test, etc.)

Quick Start:
  spark-cli create workspace ./my-project
  cd my-project
  spark-cli use AppModel
  spark-cli use AppAPI
  spark-cli sync
  cd AppAPI && spark-cli run build`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("spark-cli %s (%s %s)\n", Version, Commit, Date))
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
