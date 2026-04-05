package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newSecretMappingCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <resourceFilePath>",
		Short: "Create a Flink secret mapping.",
		Long:  "Create a Flink environment secret mapping in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.secretMappingCreate,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) secretMappingCreate(cmd *cobra.Command, args []string) error {
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

	sdkOutputMapping, err := client.CreateSecretMapping(c.createContext(), environment, sdkMapping)
	if err != nil {
		return err
	}

	return printSecretMappingOutput(cmd, sdkOutputMapping)
}
