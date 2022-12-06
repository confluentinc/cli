package ksql

import (
	"context"
	"fmt"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	listFields           = []string{"Id", "Name", "OutputTopicPrefix", "KafkaClusterId", "Storage", "Endpoint", "Status", "DetailedProcessingLog"}
	listHumanLabels      = []string{"ID", "Name", "Topic Prefix", "Kafka", "Storage", "Endpoint", "Status", "Detailed Processing Log"}
	listStructuredLabels = []string{"id", "name", "topic_prefix", "kafka", "storage", "endpoint", "status", "detailed_processing_log"}
)

func (c *ksqlCommand) newListCommand(resource string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List ksqlDB %ss.", resource),
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ksqlCommand) list(cmd *cobra.Command, _ []string) error {
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId()}
	clusters, err := c.PrivateClient.KSQL.List(context.Background(), req)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		outputWriter.AddElement(c.updateKsqlClusterForDescribeAndList(cluster))
	}
	return outputWriter.Out()
}
