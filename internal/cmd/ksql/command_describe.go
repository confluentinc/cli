package ksql

import (
	"fmt"
	"os"

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
	cluster, _, err := c.V2Client.DescribeKsqlCluster(args[0])
	if err != nil {
		return errors.CatchKSQLNotFoundError(err, args[0])
	}

	if isApp {
		_, _ = fmt.Fprintln(os.Stderr, errors.KSQLAppDeprecateWarning)
	}
	// TODO bring back formatting
	//return output.DescribeObject(cmd, c.updateKsqlClusterForDescribeAndList(cluster)), describeFields, describeHumanRenames, describeStructuredRenames)
	return output.DescribeObject(cmd, cluster, describeFields, describeHumanRenames, describeStructuredRenames)
}
