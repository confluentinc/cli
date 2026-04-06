package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newSecretMappingDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink secret mapping.",
		Long:  "Describe a Flink environment secret mapping in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.secretMappingDescribe,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) secretMappingDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkOutputMapping, err := client.DescribeSecretMapping(c.createContext(), environment, name)
	if err != nil {
		return err
	}

	return printSecretMappingOutput(cmd, sdkOutputMapping)
}
