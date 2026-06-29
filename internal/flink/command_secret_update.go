package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/resource"
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

	// Block the update if the resource is CFK-owned; if unreadable, let the update report it.
	if existingSecret, describeErr := client.DescribeSecret(c.createContext(), secretName); describeErr == nil {
		if err := errIfCfkManaged(resource.FlinkSecret, secretName, existingSecret.Metadata.GetAnnotations()); err != nil {
			return err
		}
	}

	sdkOutputSecret, err := client.UpdateSecret(c.createContext(), secretName, sdkSecret)
	if err != nil {
		return err
	}

	return printSecretOutput(cmd, sdkOutputSecret)
}
