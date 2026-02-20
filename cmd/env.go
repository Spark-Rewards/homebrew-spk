package cmd

import (
	"fmt"
	"strings"

	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage workspace environment variables",
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspace environment variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		ws, err := workspace.Load(wsPath)
		if err != nil {
			return err
		}

		if len(ws.Env) == 0 {
			fmt.Println("No environment variables set")
			return nil
		}

		for k, v := range ws.Env {
			fmt.Printf("%s=%s\n", k, v)
		}
		return nil
	},
}

var envSetCmd = &cobra.Command{
	Use:   "set <KEY=VALUE> [KEY=VALUE...]",
	Short: "Set workspace environment variables",
	Long: `Sets one or more environment variables in the workspace manifest.

Example:
  spk env set NODE_ENV=development AWS_REGION=us-east-1`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		ws, err := workspace.Load(wsPath)
		if err != nil {
			return err
		}

		if ws.Env == nil {
			ws.Env = make(map[string]string)
		}

		for _, arg := range args {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid format '%s' — use KEY=VALUE", arg)
			}
			ws.Env[parts[0]] = parts[1]
			fmt.Printf("  %s=%s\n", parts[0], parts[1])
		}

		return workspace.Save(wsPath, ws)
	},
}

var envUnsetCmd = &cobra.Command{
	Use:   "unset <KEY> [KEY...]",
	Short: "Remove workspace environment variables",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		ws, err := workspace.Load(wsPath)
		if err != nil {
			return err
		}

		for _, key := range args {
			delete(ws.Env, key)
			fmt.Printf("  removed %s\n", key)
		}

		return workspace.Save(wsPath, ws)
	},
}

// ExportString returns a shell-compatible export string for all env vars
var envExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Print export statements for shell eval",
	Long: `Outputs workspace env vars as shell export statements.
Use with eval to load into your shell:

  eval $(spk env export)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		ws, err := workspace.Load(wsPath)
		if err != nil {
			return err
		}

		for k, v := range ws.Env {
			fmt.Printf("export %s=%q\n", k, v)
		}
		return nil
	},
}

func init() {
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envUnsetCmd)
	envCmd.AddCommand(envExportCmd)
	rootCmd.AddCommand(envCmd)
}
