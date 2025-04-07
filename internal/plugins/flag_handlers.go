package plugins

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// FlagHandler defines the interface for handling different flag types
type FlagHandler interface {
	// AddFlag adds a flag to the command
	AddFlag(cmd *cobra.Command, flag *Flag) error
	// ValidateValue validates the flag value
	ValidateValue(flag *Flag, value string) error
}

// Flag represents a command flag with its configuration
type Flag struct {
	Name        string
	Shorthand   string
	Type        string
	Description struct {
		IT string `yaml:"it"`
		EN string `yaml:"en"`
		ES string `yaml:"es"`
	} `yaml:"description"`
	Required    bool     `yaml:"required"`
	Default     string   `yaml:"default,omitempty"`
	ValidValues []string `yaml:"valid_values,omitempty"`
}

// StringFlagHandler handles string flags
type StringFlagHandler struct{}

func (h *StringFlagHandler) AddFlag(cmd *cobra.Command, flag *Flag) error {
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

	defaultValue := flag.Default
	if shorthand != "" {
		cmd.Flags().StringP(flagName, shorthand, defaultValue, flag.Description.EN)
	} else {
		cmd.Flags().String(flagName, defaultValue, flag.Description.EN)
	}

	if flag.Required {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			return fmt.Errorf("failed to mark flag %s as required: %w", flagName, err)
		}
	}

	return nil
}

func (h *StringFlagHandler) ValidateValue(flag *Flag, value string) error {
	if len(flag.ValidValues) > 0 {
		// If value is empty and there's a default value, use that for validation
		if value == "" && flag.Default != "" {
			value = flag.Default
		}

		validValuesMap := make(map[string]bool)
		for _, v := range flag.ValidValues {
			validValuesMap[v] = true
		}

		if !validValuesMap[value] {
			return fmt.Errorf("invalid value for flag %s: %s. Valid values are: %s",
				flag.Name, value, strings.Join(flag.ValidValues, ", "))
		}
	}
	return nil
}

// BoolFlagHandler handles boolean flags
type BoolFlagHandler struct{}

func (h *BoolFlagHandler) AddFlag(cmd *cobra.Command, flag *Flag) error {
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

	defaultValue := flag.Default == "true"
	if shorthand != "" {
		cmd.Flags().BoolP(flagName, shorthand, defaultValue, flag.Description.EN)
	} else {
		cmd.Flags().Bool(flagName, defaultValue, flag.Description.EN)
	}

	if flag.Required {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			return fmt.Errorf("failed to mark flag %s as required: %w", flagName, err)
		}
	}

	return nil
}

func (h *BoolFlagHandler) ValidateValue(flag *Flag, value string) error {
	if len(flag.ValidValues) > 0 {
		// If value is empty and there's a default value, use that for validation
		if value == "" && flag.Default != "" {
			value = flag.Default
		}

		validValuesMap := make(map[string]bool)
		for _, v := range flag.ValidValues {
			validValuesMap[v] = true
		}

		if !validValuesMap[value] {
			return fmt.Errorf("invalid value for flag %s: %s. Valid values are: %s",
				flag.Name, value, strings.Join(flag.ValidValues, ", "))
		}
	}
	return nil
}

// IntFlagHandler handles integer flags
type IntFlagHandler struct{}

func (h *IntFlagHandler) AddFlag(cmd *cobra.Command, flag *Flag) error {
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

	defaultValue := 0
	if flag.Default != "" {
		if _, err := fmt.Sscanf(flag.Default, "%d", &defaultValue); err != nil {
			return fmt.Errorf("invalid default value for int flag %s: %w", flagName, err)
		}
	}

	if shorthand != "" {
		cmd.Flags().IntP(flagName, shorthand, defaultValue, flag.Description.EN)
	} else {
		cmd.Flags().Int(flagName, defaultValue, flag.Description.EN)
	}

	if flag.Required {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			return fmt.Errorf("failed to mark flag %s as required: %w", flagName, err)
		}
	}

	return nil
}

func (h *IntFlagHandler) ValidateValue(flag *Flag, value string) error {
	if len(flag.ValidValues) > 0 {
		// If value is empty and there's a default value, use that for validation
		if value == "" && flag.Default != "" {
			value = flag.Default
		}

		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err != nil {
			return fmt.Errorf("invalid integer value for flag %s: %s", flag.Name, value)
		}

		validValuesMap := make(map[int]bool)
		for _, v := range flag.ValidValues {
			var intValidValue int
			if _, err := fmt.Sscanf(v, "%d", &intValidValue); err != nil {
				return fmt.Errorf("invalid valid value for int flag %s: %s", flag.Name, v)
			}
			validValuesMap[intValidValue] = true
		}

		if !validValuesMap[intValue] {
			return fmt.Errorf("invalid value for flag %s: %d. Valid values are: %s",
				flag.Name, intValue, strings.Join(flag.ValidValues, ", "))
		}
	}
	return nil
}

// EnumFlagHandler handles enum flags (flags with valid_values)
type EnumFlagHandler struct{}

func (h *EnumFlagHandler) AddFlag(cmd *cobra.Command, flag *Flag) error {
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

	defaultValue := flag.Default
	if shorthand != "" {
		cmd.Flags().StringP(flagName, shorthand, defaultValue, flag.Description.EN)
	} else {
		cmd.Flags().String(flagName, defaultValue, flag.Description.EN)
	}

	if flag.Required {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			return fmt.Errorf("failed to mark flag %s as required: %w", flagName, err)
		}
	}

	return nil
}

func (h *EnumFlagHandler) ValidateValue(flag *Flag, value string) error {
	// If value is empty and there's a default value, use that for validation
	if value == "" && flag.Default != "" {
		value = flag.Default
	}

	// Always validate against valid values for enum flags
	validValuesMap := make(map[string]bool)
	for _, v := range flag.ValidValues {
		validValuesMap[v] = true
	}

	if !validValuesMap[value] {
		return fmt.Errorf("invalid value for flag %s: %s. Valid values are: %s",
			flag.Name, value, strings.Join(flag.ValidValues, ", "))
	}
	return nil
}

// FlagHandlers maps flag types to their handlers
var FlagHandlers = map[string]FlagHandler{
	"string": &StringFlagHandler{},
	"bool":   &BoolFlagHandler{},
	"int":    &IntFlagHandler{},
	"enum":   &EnumFlagHandler{},
}

// GetFlagHandler returns the appropriate handler for a flag type
func GetFlagHandler(flagType string, flag *Flag) (FlagHandler, error) {
	// If the flag has valid values, treat it as an enum regardless of its type
	if len(flag.ValidValues) > 0 {
		return FlagHandlers["enum"], nil
	}

	handler, exists := FlagHandlers[flagType]
	if !exists {
		return nil, fmt.Errorf("unsupported flag type: %s", flagType)
	}
	return handler, nil
}
