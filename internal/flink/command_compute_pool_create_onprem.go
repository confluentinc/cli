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

func (c *command) newComputePoolCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create <resourceFilePath>",
		Short:       "Create a Flink compute pool in Confluent Platform.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.computePoolCreateOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) computePoolCreateOnPrem(cmd *cobra.Command, args []string) error {
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

	var computePool cmfsdk.ComputePool
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &computePool)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &computePool)
	default:
		return errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
	}
	if err != nil {
		return err
	}

	outputComputePool, err := client.CreateComputePool(c.createContext(), environment, computePool)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)

		// nil pointer handling for creation timestamp
		var creationTime string
		if outputComputePool.GetMetadata().CreationTimestamp != nil {
			creationTime = *outputComputePool.GetMetadata().CreationTimestamp
		} else {
			creationTime = ""
		}

		table.Add(&computePoolOutOnPrem{
			CreationTime: creationTime,
			Name:         computePool.GetMetadata().Name,
			Type:         computePool.GetSpec().Type,
			Phase:        computePool.GetStatus().Phase,
		})
		return table.Print()
	}

	return output.SerializedOutput(cmd, outputComputePool)
}
