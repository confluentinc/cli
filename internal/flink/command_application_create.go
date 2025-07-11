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

	resourceFilePath := args[0]
	// Read file contents
	data, err := os.ReadFile(resourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var application cmfsdk.FlinkApplication
	var localApp localFlinkApplication
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &application)
	case ".yaml", ".yml":
		// Unmarshal into local struct with YAML tags
		if err = yaml.Unmarshal(data, &localApp); err != nil {
			return fmt.Errorf("failed to unmarshal YAML: %w", err)
		}

		// Convert to JSON bytes to use the SDK struct's JSON tags
		jsonBytes, err := json.Marshal(localApp)
		if err != nil {
			return fmt.Errorf("failed to convert to JSON: %v", err)
		}

		// Now unmarshal into the SDK struct using JSON unmarshaler
		if err = json.Unmarshal(jsonBytes, &application); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %v", err)
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

	if output.GetFormat(cmd) == output.YAML {
		// Convert the outputApplication to our local struct for correct YAML field names
		jsonBytes, err := json.Marshal(outputApplication)
		if err != nil {
			return err
		}
		var outputLocalApp localFlinkApplication
		if err = json.Unmarshal(jsonBytes, &outputLocalApp); err != nil {
			return err
		}
		// Output the local struct for correct YAML field names
		out, err := yaml.Marshal(outputLocalApp)
		if err != nil {
			return err
		}
		output.Print(false, string(out))
		return nil
	}

	return output.SerializedOutput(cmd, outputApplication)
}
