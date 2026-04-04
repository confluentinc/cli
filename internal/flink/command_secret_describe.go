package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newSecretDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink secret in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.secretDescribe,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) secretDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkOutputSecret, err := client.DescribeSecret(c.createContext(), name)
	if err != nil {
		return err
	}

	return printSecretOutput(cmd, sdkOutputSecret)
}
