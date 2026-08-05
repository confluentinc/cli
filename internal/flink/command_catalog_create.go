package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newCatalogCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <resourceFilePath>",
		Short: "Create a Flink catalog in Confluent Platform.",
		Long:  "Create a Flink catalog in Confluent Platform that provides metadata about tables and other database objects such as views and functions.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.catalogCreate,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) catalogCreate(cmd *cobra.Command, args []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkCatalog, err := readCatalogResourceFile(args[0])
	if err != nil {
		return err
	}

	sdkOutputCatalog, err := client.CreateCatalog(c.createContext(), sdkCatalog)
	if err != nil {
		return err
	}

	return printCatalogOutput(cmd, sdkOutputCatalog)
}
