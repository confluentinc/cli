package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCatalogDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink Catalog in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.catalogDescribe,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) catalogDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	outputCatalog, err := client.DescribeCatalog(c.createContext(), name)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)

		// Populate the databases field with the names of the databases
		databases := make([]string, 0, len(outputCatalog.Spec.KafkaClusters))
		for _, kafkaCluster := range outputCatalog.Spec.KafkaClusters {
			databases = append(databases, kafkaCluster.DatabaseName)
		}

		table.Add(&catalogOut{
			CreationTime: outputCatalog.Metadata.GetCreationTimestamp(),
			Name:         outputCatalog.Metadata.Name,
			Databases:    databases,
		})
		return table.Print()
	}

	return output.SerializedOutput(cmd, outputCatalog)
}
