package flink

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newApplicationCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <resourceFilePath>",
		Short: "Create a Flink application.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationCreate,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlagWithHumanRestricted(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) applicationCreate(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Check if the application already exists
	resourceFilePath := args[0]
	// Read file contents
	data, err := os.ReadFile(resourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var application cmfsdk.FlinkApplication
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &application)
	case ".yaml", ".yml":
		// First unmarshal into a generic map to preserve the case
		var raw map[string]interface{}
		if err = yaml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("failed to unmarshal YAML: %v", err)
		}

		// Convert to JSON bytes to use the struct's JSON tags
		jsonBytes, err := json.Marshal(raw)
		if err != nil {
			return fmt.Errorf("failed to convert to JSON: %v", err)
		}

		// Now unmarshal into the struct using JSON unmarshaler
		// This will respect the json tags in the struct
		if err = json.Unmarshal(jsonBytes, &application); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %v", err)
		}

		// Verify the apiVersion was set correctly
		if application.ApiVersion == "" {
			// If still empty, try to get it directly from the raw map
			if apiVer, ok := raw["apiVersion"].(string); ok {
				application.ApiVersion = apiVer
			}
		}

	default:
		return errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
	}
	if err != nil {
		return err
	}

	outputApplication, err := client.CreateApplication(c.createContext(), environment, application)
	if err != nil {
		return err
	}

	return output.SerializedOutput(cmd, outputApplication)
}
