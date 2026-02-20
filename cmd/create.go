package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Spark-Rewards/homebrew-spark-cli/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	createAWSProfile string
	createAWSRegion  string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources (workspace)",
}

var createWorkspaceCmd = &cobra.Command{
	Use:   "workspace [path]",
	Short: "Create a new spark-cli workspace",
	Long: `Creates a new workspace directory with a .spk/workspace.json manifest.
If the directory doesn't exist, it will be created.

Examples:
  spark-cli create workspace .                     # current dir
  spark-cli create workspace ./my-project          # relative path
  spark-cli create workspace ~/Projects/my-app     # absolute path`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := args[0]

		absPath, err := filepath.Abs(targetPath)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		// Create directory if it doesn't exist
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Check if workspace already exists
		manifestPath := workspace.ManifestPath(absPath)
		if _, err := os.Stat(manifestPath); err == nil {
			return fmt.Errorf("workspace already exists at %s", absPath)
		}

		name := filepath.Base(absPath)

		ws, err := workspace.Create(absPath, name, createAWSProfile, createAWSRegion)
		if err != nil {
			return err
		}

		if err := workspace.GenerateVSCodeWorkspace(absPath); err != nil {
			fmt.Printf("Warning: failed to create VS Code workspace: %v\n", err)
		}

		fmt.Printf("Workspace '%s' created at %s\n", ws.Name, absPath)
		fmt.Printf("  VS Code:     %s\n", workspace.VSCodeWorkspacePath(absPath))
		if ws.AWSProfile != "" {
			fmt.Printf("  AWS Profile: %s\n", ws.AWSProfile)
		}
		if ws.AWSRegion != "" {
			fmt.Printf("  AWS Region:  %s\n", ws.AWSRegion)
		}
		fmt.Println("\nNext steps:")
		fmt.Printf("  cd %s\n", absPath)
		fmt.Println("  spark-cli use <org/repo>")
		return nil
	},
}

func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func init() {
	createWorkspaceCmd.Flags().StringVar(&createAWSProfile, "aws-profile", "", "AWS SSO profile name")
	createWorkspaceCmd.Flags().StringVar(&createAWSRegion, "aws-region", "", "Default AWS region")
	createCmd.AddCommand(createWorkspaceCmd)
	rootCmd.AddCommand(createCmd)
}
