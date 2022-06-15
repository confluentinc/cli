package ksql

import (
	"context"
	"fmt"
	"os"

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

func (c *ksqlCommand) newDescribeCommand(isApp bool) *cobra.Command {
	shortText := "Describe a ksqlDB cluster."
	var longText string
	runCommand := c.describeCluster
	if isApp {
		// DEPRECATED: this line should be removed before CLI v3, this work is tracked in https://confluentinc.atlassian.net/browse/KCI-1411
		shortText = "DEPRECATED: Describe a ksqlDB app."
		longText = "DEPRECATED: Describe a ksqlDB app. " + errors.KSQLAppDeprecateWarning
		runCommand = c.describeApp
	}

	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             shortText,
		Long:              longText,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              runCommand,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ksqlCommand) describeCluster(cmd *cobra.Command, args []string) error {
	return c.describe(cmd, args, false)
}

func (c *ksqlCommand) describeApp(cmd *cobra.Command, args []string) error {
	return c.describe(cmd, args, true)
}

func (c *ksqlCommand) describe(cmd *cobra.Command, args []string, isApp bool) error {
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId(), Id: args[0]}
	cluster, err := c.Client.KSQL.Describe(context.Background(), req)
	if err != nil {
		return errors.CatchKSQLNotFoundError(err, args[0])
	}

	if isApp {
		_, _ = fmt.Fprintln(os.Stderr, errors.KSQLAppDeprecateWarning)
	}
	return output.DescribeObject(cmd, c.updateKsqlClusterStatus(cluster), describeFields, describeHumanRenames, describeStructuredRenames)
}
