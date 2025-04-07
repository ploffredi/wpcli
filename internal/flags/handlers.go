package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// StringFlagHandler handles string flags
type StringFlagHandler struct{}

func (h *StringFlagHandler) AddFlag(cmd *cobra.Command, flag *Flag) error {
	flagName := NormalizeFlagName(flag.Name)
	shorthand := NormalizeShorthand(flag.Shorthand)
	defaultValue := flag.Default
	description := flag.GetDescription("en")

	if shorthand != "" {
		cmd.Flags().StringP(flagName, shorthand, defaultValue, description)
	} else {
		cmd.Flags().String(flagName, defaultValue, description)
	}

	if flag.Required {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			return fmt.Errorf("failed to mark flag %s as required: %w", flagName, err)
		}
	}

	return nil
}

func (h *StringFlagHandler) ValidateValue(flag *Flag, value string) error {
	if !flag.IsValidValue(value) {
		return fmt.Errorf("invalid value for flag %s: %s. Valid values are: %s",
			flag.Name, value, strings.Join(flag.ValidValues, ", "))
	}
	return nil
}

func (h *StringFlagHandler) GetValue(cmd *cobra.Command, flagName string) (string, error) {
	value, _ := cmd.Flags().GetString(flagName)
	return value, nil
}

// BoolFlagHandler handles boolean flags
type BoolFlagHandler struct{}

func (h *BoolFlagHandler) AddFlag(cmd *cobra.Command, flag *Flag) error {
	flagName := NormalizeFlagName(flag.Name)
	shorthand := NormalizeShorthand(flag.Shorthand)
	defaultValue := flag.Default == "true"
	description := flag.GetDescription("en")

	if shorthand != "" {
		cmd.Flags().BoolP(flagName, shorthand, defaultValue, description)
	} else {
		cmd.Flags().Bool(flagName, defaultValue, description)
	}

	if flag.Required {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			return fmt.Errorf("failed to mark flag %s as required: %w", flagName, err)
		}
	}

	return nil
}

func (h *BoolFlagHandler) ValidateValue(flag *Flag, value string) error {
	if !flag.IsValidValue(value) {
		return fmt.Errorf("invalid value for flag %s: %s. Valid values are: %s",
			flag.Name, value, strings.Join(flag.ValidValues, ", "))
	}
	return nil
}

func (h *BoolFlagHandler) GetValue(cmd *cobra.Command, flagName string) (string, error) {
	value, _ := cmd.Flags().GetBool(flagName)
	return fmt.Sprintf("%v", value), nil
}

// IntFlagHandler handles integer flags
type IntFlagHandler struct{}

func (h *IntFlagHandler) AddFlag(cmd *cobra.Command, flag *Flag) error {
	flagName := NormalizeFlagName(flag.Name)
	shorthand := NormalizeShorthand(flag.Shorthand)
	defaultValue := 0
	if flag.Default != "" {
		if _, err := fmt.Sscanf(flag.Default, "%d", &defaultValue); err != nil {
			return fmt.Errorf("invalid default value for int flag %s: %w", flagName, err)
		}
	}

	description := flag.GetDescription("en")

	if shorthand != "" {
		cmd.Flags().IntP(flagName, shorthand, defaultValue, description)
	} else {
		cmd.Flags().Int(flagName, defaultValue, description)
	}

	if flag.Required {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			return fmt.Errorf("failed to mark flag %s as required: %w", flagName, err)
		}
	}

	return nil
}

func (h *IntFlagHandler) ValidateValue(flag *Flag, value string) error {
	var intValue int
	if _, err := fmt.Sscanf(value, "%d", &intValue); err != nil {
		return fmt.Errorf("invalid integer value for flag %s: %s", flag.Name, value)
	}

	if !flag.IsValidValue(value) {
		return fmt.Errorf("invalid value for flag %s: %d. Valid values are: %s",
			flag.Name, intValue, strings.Join(flag.ValidValues, ", "))
	}
	return nil
}

func (h *IntFlagHandler) GetValue(cmd *cobra.Command, flagName string) (string, error) {
	value, _ := cmd.Flags().GetInt(flagName)
	return fmt.Sprintf("%d", value), nil
}

// EnumFlagHandler handles enum flags
type EnumFlagHandler struct{}

func (h *EnumFlagHandler) AddFlag(cmd *cobra.Command, flag *Flag) error {
	flagName := NormalizeFlagName(flag.Name)
	shorthand := NormalizeShorthand(flag.Shorthand)
	defaultValue := flag.Default
	description := flag.GetDescription("en")

	if len(flag.ValidValues) > 0 {
		description = fmt.Sprintf("%s (valid values: %s)", description, strings.Join(flag.ValidValues, ", "))
	}

	if shorthand != "" {
		cmd.Flags().StringP(flagName, shorthand, defaultValue, description)
	} else {
		cmd.Flags().String(flagName, defaultValue, description)
	}

	if flag.Required {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			return fmt.Errorf("failed to mark flag %s as required: %w", flagName, err)
		}
	}

	return nil
}

func (h *EnumFlagHandler) ValidateValue(flag *Flag, value string) error {
	// Only validate if valid values are defined
	if len(flag.ValidValues) > 0 {
		// If value is empty and there's a default value, use that for validation
		if value == "" && flag.Default != "" {
			value = flag.Default
		}

		// Check if the value is in the list of valid values
		validValuesMap := flag.GetValidValues()
		if !validValuesMap[value] {
			return fmt.Errorf("invalid value for flag %s: %s. Valid values are: %s",
				flag.Name, value, strings.Join(flag.ValidValues, ", "))
		}
	}

	return nil
}

func (h *EnumFlagHandler) GetValue(cmd *cobra.Command, flagName string) (string, error) {
	value, _ := cmd.Flags().GetString(flagName)
	return value, nil
}

// GetHandler returns the appropriate handler for a flag type
func GetHandler(flagType FlagType, flag *Flag) FlagHandler {
	// If the flag has valid values, treat it as an enum regardless of its type
	if flagType == TypeEnum || (flagType == TypeString && len(flag.ValidValues) > 0) {
		return &EnumFlagHandler{}
	}

	switch flagType {
	case TypeString:
		return &StringFlagHandler{}
	case TypeBool:
		return &BoolFlagHandler{}
	case TypeInt:
		return &IntFlagHandler{}
	default:
		return &StringFlagHandler{} // Default to string handler
	}
}
