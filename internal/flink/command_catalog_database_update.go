package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/resource"
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

	// Block the update if the resource is CFK-owned; if unreadable, let the update report it.
	if existingDatabase, describeErr := client.DescribeDatabase(c.createContext(), catalogName, databaseName); describeErr == nil {
		if err := errIfCfkManaged(resource.FlinkDatabase, databaseName, existingDatabase.Metadata.GetAnnotations()); err != nil {
			return err
		}
	}

	if err := client.UpdateDatabase(c.createContext(), catalogName, databaseName, sdkDatabase); err != nil {
		return err
	}

	sdkOutputDatabase, err := client.DescribeDatabase(c.createContext(), catalogName, databaseName)
	if err != nil {
		return fmt.Errorf("database %q was updated successfully, but failed to retrieve updated details: %w", databaseName, err)
	}

	return printDatabaseOutput(cmd, sdkOutputDatabase, catalogName)
}
