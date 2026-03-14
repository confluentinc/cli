package mcp

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ValidateParams validates skill parameters against CLI flag definitions.
// This implements requirement EXEC-07: parameter validation before execution.
//
// Returns an error if:
//   - A parameter name doesn't match any registered flag (unknown parameter)
//   - A required flag is missing and has no default value
//
// Parameters:
//   - cmd: The cobra command with registered flags
//   - params: Map of parameter names to values from skill invocation
func ValidateParams(cmd *cobra.Command, params map[string]interface{}) error {
	// Check for unknown parameters
	for paramName := range params {
		flag := cmd.Flags().Lookup(paramName)
		if flag == nil {
			return fmt.Errorf("unknown parameter: %s", paramName)
		}
	}

	// Check for missing required flags
	var missingRequired []string
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		// Skip if not required
		annotations := flag.Annotations
		if annotations == nil {
			return
		}

		// Check for BashCompOneRequiredFlag annotation (used by CLI for required flags)
		if _, required := annotations["cobra_annotation_bash_completion_one_required_flag"]; !required {
			return
		}

		// Check if parameter provided
		if _, provided := params[flag.Name]; provided {
			return
		}

		// Check if flag has a non-empty default value
		if flag.DefValue != "" {
			return
		}

		missingRequired = append(missingRequired, flag.Name)
	})

	if len(missingRequired) > 0 {
		return fmt.Errorf("required parameter missing: %s", missingRequired[0])
	}

	return nil
}

// MapParamsToFlags maps skill parameters to CLI flags with type conversion.
// Validates parameters first via ValidateParams, then sets each flag value.
//
// Type handling:
//   - Converts Go types to strings via fmt.Sprint() for pflag.Set()
//   - pflag handles type parsing (booleans, integers, arrays, etc.)
//   - Skips nil values (allows using flag defaults)
//
// Returns an error if validation fails or flag.Set() fails (type mismatch).
func MapParamsToFlags(cmd *cobra.Command, params map[string]interface{}) error {
	// Validate parameters first (EXEC-07)
	if err := ValidateParams(cmd, params); err != nil {
		return err
	}

	// Set each parameter as a flag
	for name, value := range params {
		// Skip nil values - use flag default
		if value == nil {
			continue
		}

		// Convert value to string for pflag.Set()
		strValue := fmt.Sprint(value)

		// Set the flag
		if err := cmd.Flags().Set(name, strValue); err != nil {
			return fmt.Errorf("failed to set flag %s: %w", name, err)
		}
	}

	return nil
}

// ForceJSONOutput sets --output=json if the flag exists.
// This ensures skill output is machine-readable JSON for MCP server parsing.
//
// Returns nil if the --output flag doesn't exist (some commands don't support it).
// This is not an error - we only force JSON when the command supports it.
func ForceJSONOutput(cmd *cobra.Command) error {
	// Check if --output flag exists
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		// Flag doesn't exist - not an error, just skip
		return nil
	}

	// Set to json
	return cmd.Flags().Set("output", "json")
}
