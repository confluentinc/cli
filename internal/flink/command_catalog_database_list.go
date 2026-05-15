package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCatalogDatabaseListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink databases in a catalog in Confluent Platform.",
		Args:  cobra.NoArgs,
		RunE:  c.catalogDatabaseList,
	}

	cmd.Flags().String("catalog", "", "Name of the catalog.")
	cobra.CheckErr(cmd.MarkFlagRequired("catalog"))
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) catalogDatabaseList(cmd *cobra.Command, _ []string) error {
	catalogName, err := cmd.Flags().GetString("catalog")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkDatabases, err := client.ListDatabases(c.createContext(), catalogName)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, db := range sdkDatabases {
			var creationTime string
			if db.GetMetadata().CreationTimestamp != nil {
				creationTime = *db.GetMetadata().CreationTimestamp
			}
			list.Add(&databaseOut{
				CreationTime: creationTime,
				Name:         db.GetMetadata().Name,
				Catalog:      catalogName,
			})
		}
		return list.Print()
	}

	localDatabases := make([]LocalKafkaDatabase, 0, len(sdkDatabases))
	for _, sdkDatabase := range sdkDatabases {
		localDatabases = append(localDatabases, convertSdkDatabaseToLocalDatabase(sdkDatabase))
	}

	return output.SerializedOutput(cmd, localDatabases)
}
