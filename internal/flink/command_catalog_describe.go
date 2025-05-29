package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCatalogDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink catalog in Confluent Platform.",
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
		databases := make([]string, 0, len(outputCatalog.GetSpec().KafkaClusters))
		for _, kafkaCluster := range outputCatalog.GetSpec().KafkaClusters {
			databases = append(databases, kafkaCluster.DatabaseName)
		}

		// nil pointer handling for creation timestamp
		var creationTime string
		if outputCatalog.GetMetadata().CreationTimestamp != nil {
			creationTime = *outputCatalog.GetMetadata().CreationTimestamp
		} else {
			creationTime = ""
		}

		table.Add(&catalogOut{
			CreationTime: creationTime,
			Name:         outputCatalog.GetMetadata().Name,
			Databases:    databases,
		})
		return table.Print()
	}

	return output.SerializedOutput(cmd, outputCatalog)
}
