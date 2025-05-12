package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ploffredi/wpcli/internal/git"
	"github.com/ploffredi/wpcli/internal/plugins"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available plugins",
	Long:  `List all available plugins from the wpstore repository`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		basePath := filepath.Join(homeDir, ".wpcli")
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return fmt.Errorf("failed to create base directory: %w", err)
		}

		repoManager := git.NewRepoManager(basePath)
		if err := repoManager.Clone(); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}

		if err := repoManager.Pull(); err != nil {
			return fmt.Errorf("failed to pull repository: %w", err)
		}

		configManager := plugins.NewConfigManager(repoManager.GetRepoPath())
		if err := configManager.Load(); err != nil {
			return fmt.Errorf("failed to load plugins configuration: %w", err)
		}

		plugins := configManager.GetPlugins()
		if len(plugins) == 0 {
			fmt.Println("No plugins found")
			return nil
		}

		fmt.Println("Available plugins:")
		fmt.Println("-----------------")
		for _, plugin := range plugins {
			fmt.Printf("Name: %s\n", plugin.Name)
			fmt.Printf("Description: %s\n", plugin.Description)
			fmt.Printf("Latest Version: %s\n", plugin.Versions[0].Version)
			fmt.Printf("UUID: %s\n", plugin.UUID)
			fmt.Println("-----------------")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
