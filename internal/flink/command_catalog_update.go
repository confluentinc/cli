package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newCatalogUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <resourceFilePath>",
		Short: "Update a Flink catalog.",
		Long:  "Update an existing Kafka Catalog in Confluent Platform from a resource file.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.catalogUpdate,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) catalogUpdate(cmd *cobra.Command, args []string) error {
	resourceFilePath := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkCatalog, err := readCatalogResourceFile(resourceFilePath)
	if err != nil {
		return err
	}

	catalogName := sdkCatalog.Metadata.Name
	if catalogName == "" {
		return fmt.Errorf("catalog name is required: ensure the resource file contains a non-empty \"metadata.name\" field")
	}

	if err := client.UpdateCatalog(c.createContext(), catalogName, sdkCatalog); err != nil {
		return err
	}

	sdkOutputCatalog, err := client.DescribeCatalog(c.createContext(), catalogName)
	if err != nil {
		return fmt.Errorf("catalog %q was updated successfully, but failed to retrieve updated details: %w", catalogName, err)
	}

	return printCatalogOutput(cmd, sdkOutputCatalog)
}
