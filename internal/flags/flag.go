package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// FlagType represents the type of a flag
type FlagType string

const (
	TypeString FlagType = "string"
	TypeBool   FlagType = "bool"
	TypeInt    FlagType = "int"
	TypeEnum   FlagType = "enum"
)

// Flag represents a command flag with its configuration
type Flag struct {
	Name        string
	Shorthand   string
	Type        FlagType
	Description map[string]string
	Required    bool
	Default     string
	ValidValues []string `yaml:"valid_values"`
}

// FlagHandler defines the interface for handling different flag types
type FlagHandler interface {
	AddFlag(cmd *cobra.Command, flag *Flag) error
	ValidateValue(flag *Flag, value string) error
	GetValue(cmd *cobra.Command, flagName string) (string, error)
}

// FlagValue represents a validated flag value
type FlagValue struct {
	Value string
	Type  FlagType
}

// NewFlag creates a new Flag with the given configuration
func NewFlag(name, shorthand string, flagType FlagType, description map[string]string, required bool, defaultValue string, validValues []string) *Flag {
	return &Flag{
		Name:        name,
		Shorthand:   shorthand,
		Type:        flagType,
		Description: description,
		Required:    required,
		Default:     defaultValue,
		ValidValues: validValues,
	}
}

// NormalizeFlagName removes the "--" prefix from flag names
func NormalizeFlagName(name string) string {
	return strings.TrimPrefix(name, "--")
}

// NormalizeShorthand removes the "-" prefix from shorthand flags
func NormalizeShorthand(shorthand string) string {
	return strings.TrimPrefix(shorthand, "-")
}

// GetDescription returns the description in the given language, falling back to "default" if not found
func (f *Flag) GetDescription(lang string) string {
	if desc, ok := f.Description[lang]; ok {
		return desc
	}
	return f.Description["default"]
}

// Validate checks if the flag configuration is valid
func (f *Flag) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("flag name cannot be empty")
	}

	if f.Type == "" {
		return fmt.Errorf("flag type cannot be empty")
	}

	// Only validate valid values for enum flags that have them
	if f.Type == TypeEnum && len(f.ValidValues) > 0 {
		if f.Default != "" && !f.IsValidValue(f.Default) {
			return fmt.Errorf("default value %s is not in valid values for enum flag %s", f.Default, f.Name)
		}
	}

	return nil
}

// IsValidValue checks if a value is valid for this flag
func (f *Flag) IsValidValue(value string) bool {
	if len(f.ValidValues) == 0 {
		return true
	}

	// If value is empty and there's a default value, use that for validation
	if value == "" && f.Default != "" {
		value = f.Default
	}

	for _, v := range f.ValidValues {
		if v == value {
			return true
		}
	}
	return false
}

// GetValidValues returns a map of valid values for quick lookup
func (f *Flag) GetValidValues() map[string]bool {
	validValuesMap := make(map[string]bool)
	for _, v := range f.ValidValues {
		validValuesMap[v] = true
	}
	return validValuesMap
}
