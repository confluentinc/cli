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
	data, err := os.ReadFile(resourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var genericData map[string]interface{}
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &genericData)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &genericData)
	default:
		return errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
	}
	if err != nil {
		return fmt.Errorf("failed to parse input file: %w", err)
	}

	jsonBytes, err := json.Marshal(genericData)
	if err != nil {
		return fmt.Errorf("failed to marshal intermediate data: %w", err)
	}

	var sdkComputePool cmfsdk.ComputePool
	if err = json.Unmarshal(jsonBytes, &sdkComputePool); err != nil {
		return fmt.Errorf("failed to bind data to ComputePool model: %w", err)
	}

	sdkOutputComputePool, err := client.CreateComputePool(c.createContext(), environment, sdkComputePool)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		var creationTime string
		if sdkOutputComputePool.GetMetadata().CreationTimestamp != nil {
			creationTime = *sdkOutputComputePool.GetMetadata().CreationTimestamp
		} else {
			creationTime = ""
		}
		table.Add(&computePoolOutOnPrem{
			CreationTime: creationTime,
			Name:         sdkComputePool.GetMetadata().Name,
			Type:         sdkComputePool.GetSpec().Type,
			Phase:        sdkOutputComputePool.GetStatus().Phase,
		})
		return table.Print()
	}

	localPool := convertSdkComputePoolToLocalComputePool(sdkOutputComputePool)
	return output.SerializedOutput(cmd, localPool)
}
