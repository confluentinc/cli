package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newSecretListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink secrets in Confluent Platform.",
		Args:  cobra.NoArgs,
		RunE:  c.secretList,
	}

	addPageSizeFlag(cmd)
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) secretList(cmd *cobra.Command, _ []string) error {
	pageSize, err := getPageSize(cmd)
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkSecrets, err := client.ListSecrets(c.createContext(), pageSize)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, secret := range sdkSecrets {
			var creationTime string
			if secret.Metadata.CreationTimestamp != nil {
				creationTime = *secret.Metadata.CreationTimestamp
			}
			list.Add(&secretOut{
				CreationTime: creationTime,
				Name:         secret.Metadata.Name,
			})
		}
		return list.Print()
	}

	localSecrets := make([]LocalSecret, 0, len(sdkSecrets))
	for _, sdkSecret := range sdkSecrets {
		localSecrets = append(localSecrets, convertSdkSecretToLocalSecret(sdkSecret))
	}

	return output.SerializedOutput(cmd, localSecrets)
}
