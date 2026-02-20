package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables",
	Long: `Manage workspace environment variables. Run without subcommand to show current values.

Examples:
  spark-cli env                           # show current env
  spark-cli env set KEY=VALUE             # set a variable
  spark-cli env link                      # symlink .env to all repos`,
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		globalEnv, _ := workspace.ReadGlobalEnv(wsPath)
		if len(globalEnv) == 0 {
			fmt.Println("No environment variables set")
			fmt.Println("Run 'spark-cli sync' to fetch credentials from AWS")
			return nil
		}

		for k, v := range globalEnv {
			display := v
			if len(display) > 50 {
				display = display[:47] + "..."
			}
			fmt.Printf("%s=%s\n", k, display)
		}
		return nil
	},
}

var envSetCmd = &cobra.Command{
	Use:   "set <KEY=VALUE> [KEY=VALUE...]",
	Short: "Set environment variables",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		newVars := make(map[string]string)
		for _, arg := range args {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid format '%s' — use KEY=VALUE", arg)
			}
			newVars[parts[0]] = parts[1]
			fmt.Printf("%s=%s\n", parts[0], parts[1])
		}

		return workspace.WriteGlobalEnv(wsPath, newVars)
	},
}

var envExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Print export statements for shell",
	Long: `Outputs env vars as shell export statements.

Usage:
  eval $(spark-cli env export)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		globalEnv, _ := workspace.ReadGlobalEnv(wsPath)
		for k, v := range globalEnv {
			fmt.Printf("export %s=%q\n", k, v)
		}
		return nil
	},
}

var envLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Symlink .env to all repos",
	Long: `Creates symlinks from each repo's .env to the workspace's global .env file.
This allows all repos to share the same environment variables.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		ws, err := workspace.Load(wsPath)
		if err != nil {
			return err
		}

		globalEnvPath := workspace.GlobalEnvPath(wsPath)

		if _, err := os.Stat(globalEnvPath); os.IsNotExist(err) {
			if err := os.WriteFile(globalEnvPath, []byte(""), 0644); err != nil {
				return fmt.Errorf("failed to create .env: %w", err)
			}
		}

		var linked int
		for name, repo := range ws.Repos {
			repoDir := filepath.Join(wsPath, repo.Path)
			if _, err := os.Stat(repoDir); os.IsNotExist(err) {
				continue
			}

			repoEnvPath := filepath.Join(repoDir, ".env")

			info, err := os.Lstat(repoEnvPath)
			if err == nil {
				if info.Mode()&os.ModeSymlink != 0 {
					os.Remove(repoEnvPath)
				} else {
					fmt.Printf("[skip] %s — .env exists (not a symlink)\n", name)
					continue
				}
			}

			relPath, _ := filepath.Rel(repoDir, globalEnvPath)
			if err := os.Symlink(relPath, repoEnvPath); err != nil {
				fmt.Printf("[fail] %s — %v\n", name, err)
				continue
			}

			fmt.Printf("[ok]   %s\n", name)
			linked++
		}

		fmt.Printf("\n%d repo(s) linked to workspace .env\n", linked)
		return nil
	},
}

func init() {
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envExportCmd)
	envCmd.AddCommand(envLinkCmd)
	rootCmd.AddCommand(envCmd)
}
