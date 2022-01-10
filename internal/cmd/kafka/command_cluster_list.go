package kafka

import (
	"context"
	"fmt"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	pkafka "github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	listFields           = []string{"Id", "Name", "Type", "ServiceProvider", "Region", "Availability", "Status"}
	listHumanLabels      = []string{"Id", "Name", "Type", "Provider", "Region", "Availability", "Status"}
	listStructuredLabels = []string{"id", "name", "type", "provider", "region", "availability", "status"}
)

func (c *clusterCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List Kafka clusters.",
		RunE:  pcmd.NewCLIRunE(c.list),
	}

	cmd.Flags().Bool("all", false, "List clusters across all environments.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) list(cmd *cobra.Command, _ []string) error {
	listAllClusters, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	var clusters []*schedv1.KafkaCluster
	if listAllClusters {
		environments, err := c.Client.Account.List(context.Background(), &orgv1.Account{})
		if err != nil {
			return err
		}

		for _, env := range environments {
			clustersOfEnv, err := pkafka.ListKafkaClusters(c.Client, env.Id)
			if err != nil {
				return err
			}
			clusters = append(clusters, clustersOfEnv...)
		}
	} else {
		clusters, err = pkafka.ListKafkaClusters(c.Client, c.EnvironmentId())
		if err != nil {
			return err
		}
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		// Add '*' only in the case where we are printing out tables
		if outputWriter.GetOutputFormat() == output.Human {
			if cluster.Id == c.Context.KafkaClusterContext.GetActiveKafkaClusterId() {
				cluster.Id = fmt.Sprintf("* %s", cluster.Id)
			} else {
				cluster.Id = fmt.Sprintf("  %s", cluster.Id)
			}
		}
		outputWriter.AddElement(convertClusterToDescribeStruct(cluster))
	}

	return outputWriter.Out()
}
