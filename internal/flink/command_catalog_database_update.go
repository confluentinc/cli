package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCatalogDatabaseUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <resourceFilePath>",
		Short: "Update a Flink database.",
		Long:  "Update a Flink database in a catalog in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.catalogDatabaseUpdate,
	}

	cmd.Flags().String("catalog", "", "Name of the catalog.")
	cobra.CheckErr(cmd.MarkFlagRequired("catalog"))
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) catalogDatabaseUpdate(cmd *cobra.Command, args []string) error {
	resourceFilePath := args[0]

	catalogName, err := cmd.Flags().GetString("catalog")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkDatabase, err := readDatabaseResourceFile(resourceFilePath)
	if err != nil {
		return err
	}

	databaseName := sdkDatabase.Metadata.Name

	if err := client.UpdateDatabase(c.createContext(), catalogName, databaseName, sdkDatabase); err != nil {
		return err
	}

	sdkOutputDatabase, err := client.DescribeDatabase(c.createContext(), catalogName, databaseName)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		var creationTime string
		if sdkOutputDatabase.GetMetadata().CreationTimestamp != nil {
			creationTime = *sdkOutputDatabase.GetMetadata().CreationTimestamp
		}
		table.Add(&databaseOut{
			CreationTime: creationTime,
			Name:         sdkOutputDatabase.GetMetadata().Name,
			Catalog:      catalogName,
		})
		return table.Print()
	}

	localDatabase := convertSdkDatabaseToLocalDatabase(sdkOutputDatabase)
	return output.SerializedOutput(cmd, localDatabase)
}
