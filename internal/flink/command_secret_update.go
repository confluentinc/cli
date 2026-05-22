package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newSecretUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <resourceFilePath>",
		Short: "Update a Flink secret.",
		Long:  "Update a Flink secret in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.secretUpdate,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) secretUpdate(cmd *cobra.Command, args []string) error {
	resourceFilePath := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkSecret, err := readSecretResourceFile(resourceFilePath)
	if err != nil {
		return err
	}

	secretName := sdkSecret.Metadata.Name
	if secretName == "" {
		return fmt.Errorf(`secret name is required: ensure the resource file contains a non-empty "metadata.name" field`)
	}

	sdkOutputSecret, err := client.UpdateSecret(c.createContext(), secretName, sdkSecret)
	if err != nil {
		return err
	}

	return printSecretOutput(cmd, sdkOutputSecret)
}
