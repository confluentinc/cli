package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newSecretMappingListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink secret mappings.",
		Long:  "List Flink environment secret mappings in Confluent Platform.",
		Args:  cobra.NoArgs,
		RunE:  c.secretMappingList,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) secretMappingList(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkMappings, err := client.ListSecretMappings(c.createContext(), environment)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, mapping := range sdkMappings {
			var creationTime, name, secretName string
			if mapping.Metadata != nil {
				if mapping.Metadata.CreationTimestamp != nil {
					creationTime = *mapping.Metadata.CreationTimestamp
				}
				if mapping.Metadata.Name != nil {
					name = *mapping.Metadata.Name
				}
			}
			if mapping.Spec != nil {
				secretName = mapping.Spec.SecretName
			}
			list.Add(&secretMappingOut{
				CreationTime: creationTime,
				Name:         name,
				SecretName:   secretName,
			})
		}
		return list.Print()
	}

	localMappings := make([]LocalSecretMapping, 0, len(sdkMappings))
	for _, sdkMapping := range sdkMappings {
		localMappings = append(localMappings, convertSdkSecretMappingToLocalSecretMapping(sdkMapping))
	}

	return output.SerializedOutput(cmd, localMappings)
}
