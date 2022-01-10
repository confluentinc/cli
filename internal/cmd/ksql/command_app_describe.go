package ksql

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	describeFields            = []string{"Id", "Name", "OutputTopicPrefix", "KafkaClusterId", "Storage", "Endpoint", "Status"}
	describeHumanRenames      = map[string]string{"KafkaClusterId": "Kafka", "OutputTopicPrefix": "Topic Prefix"}
	describeStructuredRenames = map[string]string{"KafkaClusterId": "kafka", "OutputTopicPrefix": "topic_prefix"}
)

func (c *appCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a ksqlDB app.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.describe),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *appCommand) describe(cmd *cobra.Command, args []string) error {
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId(), Id: args[0]}
	cluster, err := c.Client.KSQL.Describe(context.Background(), req)
	if err != nil {
		return errors.CatchKSQLNotFoundError(err, args[0])
	}

	return output.DescribeObject(cmd, c.updateKsqlClusterStatus(cluster), describeFields, describeHumanRenames, describeStructuredRenames)
}
