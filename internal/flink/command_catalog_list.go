package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCatalogListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink Catalogs in Confluent Platform.",
		Args:  cobra.NoArgs,
		RunE:  c.catalogList,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlagWithHumanRestricted(cmd)

	return cmd
}

func (c *command) catalogList(cmd *cobra.Command, _ []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	catalogs, err := client.ListCatalog(c.createContext())
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, catalog := range catalogs {
			list.Add(&catalogOut{
				Name:         catalog.Metadata.Name,
				ID:           catalog.Metadata.GetUid(),
				CreationTime: catalog.Metadata.GetCreationTimestamp(),
			})
		}
		return list.Print()
	}

	return output.SerializedOutput(cmd, catalogs)
}
