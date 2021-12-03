package ksql

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	listFields           = []string{"Id", "Name", "OutputTopicPrefix", "KafkaClusterId", "Storage", "Endpoint", "Status"}
	listHumanLabels      = []string{"ID", "Name", "Topic Prefix", "Kafka", "Storage", "Endpoint", "Status"}
	listStructuredLabels = []string{"id", "name", "topic_prefix", "kafka", "storage", "endpoint", "status"}
)

func (c *appCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ksqlDB apps.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}

	output.AddFlag(cmd)

	return cmd
}

func (c *appCommand) list(cmd *cobra.Command, _ []string) error {
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId()}
	clusters, err := c.Client.KSQL.List(context.Background(), req)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		outputWriter.AddElement(c.updateKsqlClusterStatus(cluster))
	}
	return outputWriter.Out()
}
