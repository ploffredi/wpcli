package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

// PluginCommandConfig represents the command configuration in the plugin's YAML file
type PluginCommandConfig struct {
	Name        string `yaml:"name"`
	Description struct {
		IT string `yaml:"it"`
		EN string `yaml:"en"`
		ES string `yaml:"es"`
	} `yaml:"description"`
	Usage    string `yaml:"usage"`
	Examples []struct {
		Command string `yaml:"command"`
	} `yaml:"examples"`
	Args []struct {
		Name        string `yaml:"name"`
		Type        string `yaml:"type"`
		Description struct {
			IT string `yaml:"it"`
			EN string `yaml:"en"`
			ES string `yaml:"es"`
		} `yaml:"description"`
		Required bool `yaml:"required"`
	} `yaml:"args"`
	Flags []struct {
		Name        string `yaml:"name"`
		Shorthand   string `yaml:"shorthand"`
		Type        string `yaml:"type"`
		Description struct {
			IT string `yaml:"it"`
			EN string `yaml:"en"`
			ES string `yaml:"es"`
		} `yaml:"description"`
		Required bool   `yaml:"required"`
		Default  string `yaml:"default,omitempty"`
	} `yaml:"flags"`
}

// PluginYAMLConfig represents the structure of a plugin's YAML configuration file
type PluginYAMLConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description struct {
		IT string `yaml:"it"`
		EN string `yaml:"en"`
		ES string `yaml:"es"`
	} `yaml:"description"`
	Commands []PluginCommandConfig `yaml:"commands"`
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
			if strings.HasPrefix(usage, "wpcli ") {
				usage = usage[6:] // Remove "wpcli " prefix
			}

			cmd := &cobra.Command{
				Use:   usage,
				Short: cmdConfigCopy.Description.EN,
				Long:  cmdConfigCopy.Description.EN,
				Args: func(cmd *cobra.Command, args []string) error {
					if len(args) < requiredArgs {
						return fmt.Errorf("requires at least %d argument(s)", requiredArgs)
					}
					return nil
				},
				RunE: func(cmd *cobra.Command, args []string) error {
					// Build command summary
					cmdStr := fmt.Sprintf("%s %s", cmdName, strings.Join(args, " "))

					// Add flags
					cmd.Flags().Visit(func(f *pflag.Flag) {
						if f.Value.Type() == "bool" {
							cmdStr += fmt.Sprintf(" --%s", f.Name)
						} else {
							cmdStr += fmt.Sprintf(" --%s=%s", f.Name, f.Value.String())
						}
					})

					fmt.Printf("Executing: %s\n", cmdStr)
					return nil
				},
			}

			// Add arguments
			for _, arg := range cmdConfigCopy.Args {
				cmd.Use = strings.ReplaceAll(cmd.Use, "<"+arg.Name+">", fmt.Sprintf("<%s>", arg.Name))
				cmd.Long = fmt.Sprintf("%s\n\nArguments:\n  %s (%s) - %s", cmd.Long, arg.Name, arg.Type, arg.Description.EN)
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
			for _, flag := range cmdConfigCopy.Flags {
				flagName := flag.Name
				if len(flagName) > 2 && flagName[:2] == "--" {
					flagName = flagName[2:]
				}

				shorthand := ""
				if flag.Shorthand != "" {
					shorthand = flag.Shorthand
					if len(shorthand) > 1 && shorthand[0] == '-' {
						shorthand = shorthand[1:]
					}
				}

				switch flag.Type {
				case "string":
					var defaultValue string
					if flag.Default != "" {
						defaultValue = flag.Default
					}
					if shorthand != "" {
						cmd.Flags().StringP(flagName, shorthand, defaultValue, flag.Description.EN)
					} else {
						cmd.Flags().String(flagName, defaultValue, flag.Description.EN)
					}

					// Add validation for specific flags
					if cmd.Name() == "list" && flagName == "format" {
						cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
							format, _ := cmd.Flags().GetString("format")
							validFormats := map[string]bool{
								"json":  true,
								"yaml":  true,
								"table": true,
							}
							if !validFormats[format] {
								return fmt.Errorf("invalid format: %s. Valid formats are: json, yaml, table", format)
							}
							return nil
						}
					}
				case "bool":
					defaultValue := flag.Default == "true"
					if shorthand != "" {
						cmd.Flags().BoolP(flagName, shorthand, defaultValue, flag.Description.EN)
					} else {
						cmd.Flags().Bool(flagName, defaultValue, flag.Description.EN)
					}
				}
				if flag.Required {
					if err := cmd.MarkFlagRequired(flagName); err != nil {
						return nil, fmt.Errorf("failed to mark flag %s as required: %w", flagName, err)
					}
				}
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
