package mcp

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestValidateParamsUnknownFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("valid-flag", "", "a valid flag")

	params := map[string]interface{}{
		"unknown-flag": "value",
	}

	err := ValidateParams(cmd, params)
	require.Error(t, err, "should error on unknown flag")
	require.Contains(t, err.Error(), "unknown parameter")
}

func TestValidateParamsMissingRequired(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("required-flag", "", "a required flag")
	cmd.MarkFlagRequired("required-flag")

	params := map[string]interface{}{}

	err := ValidateParams(cmd, params)
	require.Error(t, err, "should error on missing required flag")
	require.Contains(t, err.Error(), "required parameter missing")
}

func TestValidateParamsSuccess(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("flag1", "", "flag 1")
	cmd.Flags().Int("flag2", 0, "flag 2")

	params := map[string]interface{}{
		"flag1": "value1",
		"flag2": 42,
	}

	err := ValidateParams(cmd, params)
	require.NoError(t, err, "should pass validation for valid params")
}

func TestMapParamsToFlagsString(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "name flag")

	params := map[string]interface{}{
		"name": "test-value",
	}

	err := MapParamsToFlags(cmd, params)
	require.NoError(t, err)

	// Verify flag was set
	value, err := cmd.Flags().GetString("name")
	require.NoError(t, err)
	require.Equal(t, "test-value", value)
}

func TestMapParamsToFlagsBoolean(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("enabled", false, "enabled flag")

	params := map[string]interface{}{
		"enabled": true,
	}

	err := MapParamsToFlags(cmd, params)
	require.NoError(t, err)

	// Verify flag was set
	value, err := cmd.Flags().GetBool("enabled")
	require.NoError(t, err)
	require.True(t, value)
}

func TestMapParamsToFlagsInteger(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Int("count", 0, "count flag")

	params := map[string]interface{}{
		"count": 123,
	}

	err := MapParamsToFlags(cmd, params)
	require.NoError(t, err)

	// Verify flag was set
	value, err := cmd.Flags().GetInt("count")
	require.NoError(t, err)
	require.Equal(t, 123, value)
}

func TestMapParamsToFlagsSkipsNil(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("optional", "default-value", "optional flag")

	params := map[string]interface{}{
		"optional": nil,
	}

	err := MapParamsToFlags(cmd, params)
	require.NoError(t, err)

	// Verify flag kept default value
	value, err := cmd.Flags().GetString("optional")
	require.NoError(t, err)
	require.Equal(t, "default-value", value)
}

func TestMapParamsToFlagsValidatesFirst(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("valid-flag", "", "valid flag")

	params := map[string]interface{}{
		"unknown-flag": "value",
	}

	err := MapParamsToFlags(cmd, params)
	require.Error(t, err, "should validate before setting flags")
	require.Contains(t, err.Error(), "unknown parameter")
}

func TestMapParamsToFlagsTypeMismatch(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Int("number", 0, "number flag")

	params := map[string]interface{}{
		"number": "not-a-number",
	}

	err := MapParamsToFlags(cmd, params)
	require.Error(t, err, "should error on type mismatch")
}

func TestForceJSONOutputSetsFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("output", "human", "output format")

	err := ForceJSONOutput(cmd)
	require.NoError(t, err)

	// Verify flag was set to json
	value, err := cmd.Flags().GetString("output")
	require.NoError(t, err)
	require.Equal(t, "json", value)
}

func TestForceJSONOutputMissingFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	// No --output flag

	err := ForceJSONOutput(cmd)
	require.NoError(t, err, "should not error when --output flag missing")
}
