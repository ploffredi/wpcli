package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ploffredi/wpcli/internal/git"
	"github.com/ploffredi/wpcli/internal/plugins"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [plugin-name]",
	Short: "Get detailed information about a specific plugin (builtin)",
	Long:  `Get detailed information about a specific plugin from the wpstore repository (builtin)`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]

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

		plugin, err := configManager.GetPluginByName(pluginName)
		if err != nil {
			return fmt.Errorf("failed to get plugin information: %w", err)
		}

		fmt.Printf("Plugin Information for: %s\n", plugin.Name)
		fmt.Println("-----------------")
		fmt.Println("Description:")
		fmt.Printf("  English: %s\n", plugin.Description["en"])
		fmt.Printf("  Italian: %s\n", plugin.Description["it"])
		fmt.Printf("  Spanish: %s\n", plugin.Description["es"])
		fmt.Printf("UUID: %s\n", plugin.UUID)
		fmt.Println("\nVersions:")
		for _, version := range plugin.Versions {
			fmt.Printf("  Version: %s\n", version.Version)
			fmt.Printf("    WASM: %s\n", version.Wasm)
			fmt.Printf("    Config: %s\n", version.Conf)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
