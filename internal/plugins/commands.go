package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ploffredi/wpcli/internal/flags"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// PluginCommand represents a command that can be executed by a plugin
type PluginCommand struct {
	Name        string
	Description string
	WasmFile    string
	ConfigFile  string
	Version     string
	Subcommand  string
}

// GetPluginCommands returns a list of commands available from the plugins
func GetPluginCommands(configPath string) ([]*cobra.Command, error) {
	config := &PluginConfig{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugins.yml: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse plugins.yml: %w", err)
	}

	// Group plugins by subcommand
	subcommandGroups := make(map[string]*cobra.Command)
	subcommandVersions := make(map[string]string)
	subcommandPlugins := make(map[string]string)
	var rootCommands []*cobra.Command

	for _, plugin := range config.Plugins {
		// Sort versions in descending order to get the latest version first
		versions := make([]Version, len(plugin.Versions))
		copy(versions, plugin.Versions)
		sort.Slice(versions, func(i, j int) bool {
			return versions[i].Version > versions[j].Version
		})

		// Use only the latest version
		latestVersion := versions[0]

		// Read plugin-specific YAML configuration
		pluginConfigPath := filepath.Join(filepath.Dir(configPath), plugin.UUID, latestVersion.Version, latestVersion.Conf)
		pluginConfig, err := loadPluginConfig(pluginConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load plugin config for %s: %w", plugin.Name, err)
		}

		// Get or create the parent command for plugins with subcommands
		var parentCmd *cobra.Command
		if plugin.Subcommand != "" {
			if cmd, exists := subcommandGroups[plugin.Subcommand]; exists {
				parentCmd = cmd
			} else {
				parentCmd = &cobra.Command{
					Use:   plugin.Subcommand,
					Short: fmt.Sprintf("Commands for %s plugins (from %s v%s)", plugin.Subcommand, plugin.Name, latestVersion.Version),
					Long:  fmt.Sprintf("Commands for %s plugins\n\nVersion: %s\n\nPlugin: %s", plugin.Subcommand, latestVersion.Version, plugin.Name),
				}
				subcommandGroups[plugin.Subcommand] = parentCmd
				subcommandVersions[plugin.Subcommand] = latestVersion.Version
				subcommandPlugins[plugin.Subcommand] = plugin.Name
				rootCommands = append(rootCommands, parentCmd)
			}
		}

		// Create commands for each plugin command
		for _, cmdConfig := range pluginConfig.Commands {
			// Create a copy of cmdConfig for the closure
			cmdConfigCopy := cmdConfig

			// Count required arguments
			requiredArgs := 0
			for _, arg := range cmdConfigCopy.Args {
				if arg.Required {
					requiredArgs++
				}
			}

			// Extract command name from usage pattern
			parts := strings.Fields(cmdConfigCopy.Usage)
			var cmdName string
			if len(parts) > 0 {
				cmdName = cmdConfigCopy.Name // Use the name from the config
			} else {
				cmdName = cmdConfigCopy.Name
			}

			// Build usage pattern with arguments
			usage := cmdConfigCopy.Usage
			usage = strings.TrimPrefix(usage, "wpcli ")

			description := cmdConfigCopy.Description["en"]
			if description == "" {
				description = cmdConfigCopy.Description["default"]
			}

			cmd := &cobra.Command{
				Use:   usage,
				Short: description,
				Long:  description,
				Args: func(cmd *cobra.Command, args []string) error {
					// Validate arguments
					if len(args) < requiredArgs {
						return fmt.Errorf("requires at least %d argument(s)", requiredArgs)
					}
					return nil
				},
				PreRunE: func(cmd *cobra.Command, args []string) error {
					// First validate that all required flags are provided
					if err := cmd.ValidateRequiredFlags(); err != nil {
						return err
					}

					// Then validate all flag values
					for _, flag := range cmdConfigCopy.Flags {
						handler := flags.GetHandler(flags.ParseFlagType(string(flag.Type)), flag)
						flagName := flags.NormalizeFlagName(flag.Name)
						// Always validate if the flag was set, regardless of default value
						if cmd.Flags().Changed(flagName) {
							value, err := handler.GetValue(cmd, flagName)
							if err != nil {
								return fmt.Errorf("failed to get value for flag %s: %w", flag.Name, err)
							}

							if err := handler.ValidateValue(flag, value); err != nil {
								return err
							}
						}
					}
					return nil
				},
				RunE: func(cmd *cobra.Command, args []string) error {
					// Re-run validation in RunE to ensure errors are properly propagated
					if err := cmd.ValidateRequiredFlags(); err != nil {
						return err
					}

					// Then validate all flag values again
					for _, flag := range cmdConfigCopy.Flags {
						handler := flags.GetHandler(flags.ParseFlagType(string(flag.Type)), flag)
						flagName := flags.NormalizeFlagName(flag.Name)

						// Always validate if the flag was set, regardless of default value
						if cmd.Flags().Changed(flagName) {
							value, err := handler.GetValue(cmd, flagName)
							if err != nil {
								return fmt.Errorf("failed to get value for flag %s: %w", flag.Name, err)
							}

							if err := handler.ValidateValue(flag, value); err != nil {
								return err
							}
						}
					}

					// Only print command summary if validation passes
					cmdStr := flags.BuildCommandSummary(cmdName, args, cmd)
					fmt.Printf("Executing: %s\n", cmdStr)
					return nil
				},
			}

			// Add arguments
			for _, arg := range cmdConfigCopy.Args {
				cmd.Use = strings.ReplaceAll(cmd.Use, "<"+arg.Name+">", fmt.Sprintf("<%s>", arg.Name))
				argDesc := arg.Description["en"]
				if argDesc == "" {
					argDesc = arg.Description["default"]
				}
				cmd.Long = fmt.Sprintf("%s\n\nArguments:\n  %s (%s) - %s", cmd.Long, arg.Name, arg.Type, argDesc)
			}

			// Add examples
			if len(cmdConfigCopy.Examples) > 0 {
				examples := "\n\nExamples:\n"
				for _, example := range cmdConfigCopy.Examples {
					examples += fmt.Sprintf("  %s\n", example.Command)
				}
				cmd.Long += examples
			}

			// Add flags
			if err := flags.AddFlags(cmd, cmdConfigCopy.Flags); err != nil {
				return nil, fmt.Errorf("failed to add flags: %w", err)
			}

			// Add the command to the appropriate parent
			if parentCmd != nil {
				// Add command directly to the parent command
				cmd.Short = fmt.Sprintf("%s (from %s v%s)", cmd.Short, plugin.Name, latestVersion.Version)
				parentCmd.AddCommand(cmd)
			} else {
				// For root-level commands, add version info to the description
				cmd.Short = fmt.Sprintf("%s (from %s v%s)", cmd.Short, plugin.Name, latestVersion.Version)
				cmd.Long = fmt.Sprintf("%s\n\nPlugin: %s\nVersion: %s", cmd.Long, plugin.Name, latestVersion.Version)
				rootCommands = append(rootCommands, cmd)
			}
		}
	}

	return rootCommands, nil
}

func loadPluginConfig(configPath string) (*PluginYAMLConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin config: %w", err)
	}

	config := &PluginYAMLConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse plugin config: %w", err)
	}

	return config, nil
}

// Add this function to handle invalid subcommands
func init() {
	// Override the default behavior for invalid subcommands
	cobra.OnInitialize(func() {
		// This will be called after all flags are parsed
		// We can check if the command is valid here
	})
}
