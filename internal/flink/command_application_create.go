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

	cmd.Flags().String("environment", "", "Name of the environment to delete the Flink application from.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlagWithDefaultValue(cmd, output.JSON.String())

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) applicationCreate(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	// Disallow human output for this command
	if output.GetFormat(cmd) == output.Human {
		return errors.NewErrorWithSuggestions("human output is not supported for this command", "Try using --output flag with json or yaml.\n")
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

	var application cmfsdk.Application
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &application)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &application)
	default:
		return errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
	}
	if err != nil {
		return err
	}

	outputApplication, err := client.CreateApplication(cmd.Context(), environment, application)
	if err != nil {
		return err
	}

	return output.SerializedOutput(cmd, outputApplication)
}