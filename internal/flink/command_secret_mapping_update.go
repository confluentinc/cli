package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newSecretMappingUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <resourceFilePath>",
		Short: "Update a Flink secret mapping.",
		Long:  "Update a Flink environment secret mapping in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.secretMappingUpdate,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) secretMappingUpdate(cmd *cobra.Command, args []string) error {
	resourceFilePath := args[0]

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkMapping, err := readSecretMappingResourceFile(resourceFilePath)
	if err != nil {
		return err
	}

	var mappingName string
	if sdkMapping.Metadata != nil && sdkMapping.Metadata.Name != nil {
		mappingName = *sdkMapping.Metadata.Name
	}
	if mappingName == "" {
		return fmt.Errorf(`secret mapping name is required: ensure the resource file contains a non-empty "metadata.name" field`)
	}

	sdkOutputMapping, err := client.UpdateSecretMapping(c.createContext(), environment, mappingName, sdkMapping)
	if err != nil {
		return err
	}

	return printSecretMappingOutput(cmd, sdkOutputMapping)
}
