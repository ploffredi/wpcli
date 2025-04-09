package plugins

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ploffredi/wpcli/internal/flags"
	"gopkg.in/yaml.v3"
)

type Version struct {
	Version string `yaml:"version"`
	Wasm    string `yaml:"wasm"`
	Conf    string `yaml:"conf"`
}

type Plugin struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	UUID        string                 `yaml:"uuid"`
	Versions    []Version              `yaml:"versions"`
	Subcommand  string                 `yaml:"subcommand,omitempty"`
	Version     string                 `yaml:"version,omitempty"`
	Commands    []PluginCommandConfig  `yaml:"commands,omitempty"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"` // For plugin-specific data
}

type Settings struct {
	DefaultRepository  string   `yaml:"default_repository"`
	CacheDir           string   `yaml:"cache_dir"`
	LogLevel           string   `yaml:"log_level"`
	DefaultLanguage    string   `yaml:"default_language"`
	SupportedLanguages []string `yaml:"supported_languages"`
}

type PluginConfig struct {
	Plugins  []Plugin `yaml:"plugins"`
	Settings Settings `yaml:"settings"`
}

type ConfigManager struct {
	configPath string
	config     *PluginConfig
}

func NewConfigManager(repoPath string) *ConfigManager {
	return &ConfigManager{
		configPath: filepath.Join(repoPath, "plugins.yml"),
	}
}

func (cm *ConfigManager) Load() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read plugins.yml: %w", err)
	}

	config := &PluginConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse plugins.yml: %w", err)
	}

	cm.config = config
	return nil
}

func (cm *ConfigManager) GetPlugins() []Plugin {
	if cm.config == nil {
		return []Plugin{}
	}
	return cm.config.Plugins
}

func (cm *ConfigManager) GetPluginByName(name string) (*Plugin, error) {
	if cm.config == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	for _, plugin := range cm.config.Plugins {
		if plugin.Name == name {
			return &plugin, nil
		}
	}

	return nil, fmt.Errorf("plugin %s not found", name)
}

func (cm *ConfigManager) GetSettings() *Settings {
	if cm.config == nil {
		return nil
	}
	return &cm.config.Settings
}

// PluginCommandConfig represents the configuration for a plugin command
type PluginCommandConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Usage       string `yaml:"usage"`
	Examples    []struct {
		Description string `yaml:"description"`
		Command     string `yaml:"command"`
	} `yaml:"examples"`
	Args []struct {
		Name        string `yaml:"name"`
		Type        string `yaml:"type"`
		Description string `yaml:"description"`
		Required    bool   `yaml:"required"`
	} `yaml:"args"`
	Flags []*flags.Flag `yaml:"flags"`
	// Additional fields from PluginCommand
	WasmFile   string `yaml:"wasm_file,omitempty"`
	ConfigFile string `yaml:"config_file,omitempty"`
	Version    string `yaml:"version,omitempty"`
	Subcommand string `yaml:"subcommand,omitempty"`
}

// loadPluginConfig loads a plugin's YAML configuration file
func loadPluginConfig(configPath string) (*Plugin, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin config: %w", err)
	}

	config := &Plugin{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse plugin config: %w", err)
	}

	return config, nil
}
