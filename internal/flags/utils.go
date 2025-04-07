package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// AddFlags adds multiple flags to a command
func AddFlags(cmd *cobra.Command, flags []*Flag) error {
	for _, flag := range flags {
		if err := flag.Validate(); err != nil {
			return fmt.Errorf("invalid flag configuration: %w", err)
		}

		handler := GetHandler(flag.Type, flag)
		if err := handler.AddFlag(cmd, flag); err != nil {
			return fmt.Errorf("failed to add flag %s: %w", flag.Name, err)
		}
	}
	return nil
}

// ValidateFlags validates all flags for a command
func ValidateFlags(cmd *cobra.Command, flags []*Flag) error {
	for _, flag := range flags {
		handler := GetHandler(flag.Type, flag)
		flagName := NormalizeFlagName(flag.Name)
		value, err := handler.GetValue(cmd, flagName)
		if err != nil {
			return fmt.Errorf("failed to get value for flag %s: %w", flag.Name, err)
		}

		if err := handler.ValidateValue(flag, value); err != nil {
			return err
		}
	}
	return nil
}

// GetFlagValues returns a map of flag names to their values
func GetFlagValues(cmd *cobra.Command, flags []*Flag) (map[string]string, error) {
	values := make(map[string]string)
	for _, flag := range flags {
		handler := GetHandler(flag.Type, flag)
		flagName := NormalizeFlagName(flag.Name)
		value, err := handler.GetValue(cmd, flagName)
		if err != nil {
			return nil, fmt.Errorf("failed to get value for flag %s: %w", flag.Name, err)
		}
		values[flag.Name] = value
	}
	return values, nil
}

// BuildCommandSummary builds a string representation of the command with its arguments and flags
func BuildCommandSummary(cmdName string, args []string, cmd *cobra.Command) string {
	var parts []string
	parts = append(parts, cmdName)
	parts = append(parts, args...)

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if flag.Changed {
			parts = append(parts, fmt.Sprintf("--%s=%v", flag.Name, flag.Value))
		}
	})

	return strings.Join(parts, " ")
}

// ParseFlagType converts a string to a FlagType
func ParseFlagType(typeStr string) FlagType {
	switch strings.ToLower(typeStr) {
	case "string":
		return TypeString
	case "bool":
		return TypeBool
	case "int":
		return TypeInt
	case "enum":
		return TypeEnum
	default:
		return TypeString // Default to string type
	}
}
