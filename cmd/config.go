package cmd

import (
	"fmt"

	"github.com/Spark-Rewards/homebrew-spk/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or update global spk configuration",
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Show current global configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadGlobal()
		if err != nil {
			return err
		}

		fmt.Printf("GitHub Org:   %s\n", orDefault(cfg.DefaultGithubOrg, "(not set)"))
		fmt.Printf("AWS Profile:  %s\n", orDefault(cfg.DefaultAWSProfile, "(not set)"))
		fmt.Printf("AWS Region:   %s\n", orDefault(cfg.DefaultAWSRegion, "(not set)"))
		fmt.Printf("Workspaces:   %d registered\n", len(cfg.Workspaces))

		for _, ws := range cfg.Workspaces {
			fmt.Printf("  - %s\n", ws)
		}

		return nil
	},
}

var (
	setOrg        string
	setAWSProfile string
	setAWSRegion  string
)

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set global defaults",
	Long: `Set global configuration defaults used across all workspaces.

Examples:
  spk config set --org my-github-org
  spk config set --aws-profile dev --aws-region us-east-1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if setOrg == "" && setAWSProfile == "" && setAWSRegion == "" {
			return fmt.Errorf("provide at least one flag: --org, --aws-profile, or --aws-region")
		}

		if err := config.SetDefaults(setOrg, setAWSProfile, setAWSRegion); err != nil {
			return err
		}

		fmt.Println("Global config updated")
		return nil
	},
}

func init() {
	configSetCmd.Flags().StringVar(&setOrg, "org", "", "Default GitHub organization")
	configSetCmd.Flags().StringVar(&setAWSProfile, "aws-profile", "", "Default AWS profile")
	configSetCmd.Flags().StringVar(&setAWSRegion, "aws-region", "", "Default AWS region")

	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
