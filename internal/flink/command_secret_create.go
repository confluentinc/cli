package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newSecretCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <resourceFilePath>",
		Short: "Create a Flink secret.",
		Long:  "Create a Flink secret in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.secretCreate,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) secretCreate(cmd *cobra.Command, args []string) error {
	resourceFilePath := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkSecret, err := readSecretResourceFile(resourceFilePath)
	if err != nil {
		return err
	}

	sdkOutputSecret, err := client.CreateSecret(c.createContext(), sdkSecret)
	if err != nil {
		return err
	}

	return printSecretOutput(cmd, sdkOutputSecret)
}
