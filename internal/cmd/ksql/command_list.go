package ksql

import (
	"fmt"
	"os"

	"github.com/confluentinc/cli/internal/pkg/errors"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	listFields           = []string{"Id", "Name", "OutputTopicPrefix", "KafkaClusterId", "Storage", "Endpoint", "Status", "DetailedProcessingLog"}
	listHumanLabels      = []string{"ID", "Name", "Topic Prefix", "Kafka", "Storage", "Endpoint", "Status", "Detailed Processing Log"}
	listStructuredLabels = []string{"id", "name", "topic_prefix", "kafka", "storage", "endpoint", "status", "detailed_processing_log"}
)

func (c *ksqlCommand) newListCommand(isApp bool) *cobra.Command {
	shortText := "List ksqlDB clusters."
	var longText string
	runCommand := c.listClusters
	if isApp {
		// DEPRECATED: this should be removed before CLI v3, this work is tracked in https://confluentinc.atlassian.net/browse/KCI-1411
		shortText = "DEPRECATED: List ksqlDB apps."
		longText = "DEPRECATED: List ksqlDB apps. " + errors.KSQLAppDeprecateWarning
		runCommand = c.listApps
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: shortText,
		Long:  longText,
		Args:  cobra.NoArgs,
		RunE:  runCommand,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ksqlCommand) listClusters(cmd *cobra.Command, args []string) error {
	return c.list(cmd, args, false)
}

func (c *ksqlCommand) listApps(cmd *cobra.Command, args []string) error {
	return c.list(cmd, args, true)
}

func (c *ksqlCommand) list(cmd *cobra.Command, _ []string, isApp bool) error {
	if isApp {
		_, _ = fmt.Fprintln(os.Stderr, errors.KSQLAppDeprecateWarning)
	}

	clusters, err := c.V2Client.ListKsqlClusters(c.EnvironmentId())
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}
	for _, cluster := range clusters.Data {
		outputWriter.AddElement(c.updateKsqlClusterForDescribeAndList(&cluster))
	}
	return outputWriter.Out()
}
