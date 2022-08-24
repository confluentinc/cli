package ksql

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	describeFields            = []string{"Id", "Name", "OutputTopicPrefix", "KafkaClusterId", "Storage", "Endpoint", "Status", "DetailedProcessingLog"}
	describeHumanRenames      = map[string]string{"KafkaClusterId": "Kafka", "OutputTopicPrefix": "Topic Prefix", "DetailedProcessingLog": "Detailed Processing Log"}
	describeStructuredRenames = map[string]string{"KafkaClusterId": "kafka", "OutputTopicPrefix": "topic_prefix", "DetailedProcessingLog": "detailed_processing_log"}
)

func (c *ksqlCommand) newDescribeCommand(resource string) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             fmt.Sprintf("Describe a ksqlDB %s.", resource),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ksqlCommand) describe(cmd *cobra.Command, args []string) error {
	cluster, err := c.V2Client.DescribeKsqlCluster(args[0], c.EnvironmentId())
	if err != nil {
		return errors.CatchKSQLNotFoundError(err, args[0])
	}

	return output.DescribeObject(cmd, c.formatClusterForDisplayAndList(&cluster), describeFields, describeHumanRenames, describeStructuredRenames)
}
