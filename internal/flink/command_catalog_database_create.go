package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newCatalogDatabaseCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <resourceFilePath>",
		Short: "Create a Flink database.",
		Long:  "Create a Flink database in a catalog in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.catalogDatabaseCreate,
	}

	cmd.Flags().String("catalog", "", "Name of the catalog.")
	cobra.CheckErr(cmd.MarkFlagRequired("catalog"))
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) catalogDatabaseCreate(cmd *cobra.Command, args []string) error {
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

	sdkOutputDatabase, err := client.CreateDatabase(c.createContext(), catalogName, sdkDatabase)
	if err != nil {
		return err
	}

	return printDatabaseOutput(cmd, sdkOutputDatabase, catalogName)
}
