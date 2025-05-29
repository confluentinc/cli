package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCatalogListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink catalogs in Confluent Platform.",
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
			databases := make([]string, 0, len(catalog.GetSpec().KafkaClusters))
			for _, kafkaCluster := range catalog.GetSpec().KafkaClusters {
				databases = append(databases, kafkaCluster.DatabaseName)
			}

			// nil pointer handling for creation timestamp
			var creationTime string
			if catalog.GetMetadata().CreationTimestamp != nil {
				creationTime = *catalog.GetMetadata().CreationTimestamp
			} else {
				creationTime = ""
			}

			list.Add(&catalogOut{
				CreationTime: creationTime,
				Name:         catalog.Metadata.Name,
				Databases:    databases,
			})
		}
		return list.Print()
	}

	return output.SerializedOutput(cmd, catalogs)
}
