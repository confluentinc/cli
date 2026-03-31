package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCatalogDatabaseDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink database in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.catalogDatabaseDescribe,
	}

	cmd.Flags().String("catalog", "", "Name of the catalog.")
	cobra.CheckErr(cmd.MarkFlagRequired("catalog"))
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) catalogDatabaseDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]

	catalogName, err := cmd.Flags().GetString("catalog")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkOutputDatabase, err := client.DescribeDatabase(c.createContext(), catalogName, name)
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
