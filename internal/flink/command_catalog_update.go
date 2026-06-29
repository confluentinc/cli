package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newCatalogUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <resourceFilePath>",
		Short: "Update a Flink catalog in Confluent Platform.",
		Long:  "Update an existing Flink catalog in Confluent Platform from a resource file.",
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

	// Block mutations on CFK-owned resources. Best-effort: if the current resource
	// cannot be fetched, fall through and let the update surface the real error.
	if existingCatalog, describeErr := client.DescribeCatalog(c.createContext(), catalogName); describeErr == nil {
		if err := errIfCfkManaged(resource.FlinkCatalog, catalogName, existingCatalog.Metadata.GetAnnotations()); err != nil {
			return err
		}
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
