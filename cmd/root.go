package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ploffredi/wpcli/internal/git"
	"github.com/ploffredi/wpcli/internal/plugins"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wpcli",
	Short: "WPStore CLI - A command line interface for managing WebAssembly plugins",
	Long: `WPStore CLI is a command line interface for managing WebAssembly plugins.
It provides functionality to interact with the wpstore git repository and manage plugins.yml.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no arguments are provided, show help
		if len(args) == 0 {
			return cmd.Help()
		}
		// If an invalid command is provided, show error
		return fmt.Errorf("unknown command %q for %q\nRun '%s --help' for usage", args[0], cmd.CommandPath(), cmd.CommandPath())
	},
}

func init() {
	// Load plugin commands
	if err := loadPluginCommands(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load plugin commands: %v\n", err)
	}

	// Set up command handling
	cobra.EnableCommandSorting = false
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	// Configure cobra to handle invalid subcommands
	cobra.OnInitialize(func() {
		cobra.EnablePrefixMatching = false
		cobra.EnableCommandSorting = false
	})
}

func loadPluginCommands() error {
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

	// Load plugin commands
	pluginCommands, err := plugins.GetPluginCommands(filepath.Join(repoManager.GetRepoPath(), "plugins.yml"))
	if err != nil {
		return fmt.Errorf("failed to load plugin commands: %w", err)
	}

	// Create a map of existing command names to avoid duplicates
	existingCommands := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		existingCommands[strings.Fields(cmd.Use)[0]] = true
	}

	// Add plugin commands to root command
	for _, cmd := range pluginCommands {
		// Skip if command already exists
		cmdName := strings.Fields(cmd.Use)[0]
		if existingCommands[cmdName] {
			continue
		}
		existingCommands[cmdName] = true
		rootCmd.AddCommand(cmd)
	}

	return nil
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		// Print the error message and exit with code 1 for any error
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return nil
}
