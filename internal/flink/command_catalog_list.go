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
	pcmd.AddOutputFlag(cmd)

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
			// Populate the databases field with the names of the databases
			databases := make([]string, 0, len(catalog.Spec.KafkaClusters))
			for _, kafkaCluster := range catalog.Spec.KafkaClusters {
				databases = append(databases, kafkaCluster.DatabaseName)
			}

			list.Add(&catalogOut{
				CreationTime: catalog.Metadata.GetCreationTimestamp(),
				Name:         catalog.Metadata.Name,
				Databases:    databases,
			})
		}
		return list.Print()
	}

	return output.SerializedOutput(cmd, catalogs)
}
