package ksql

import (
	"context"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"os"

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

func (c *ksqlCommand) newListCommand(isApp bool) *cobra.Command {
	shortText := "List ksqlDB clusters."
	runCommand := c.listClusters
	if isApp {
		// DEPRECATED: this line should be removed before CLI v3, this work is tracked in https://confluentinc.atlassian.net/browse/KCI-1411
		shortText = "List ksqlDB apps. " + errors.KSQLAppDeprecateWarning
		runCommand = c.listApps
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: shortText,
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(runCommand),
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
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId()}
	clusters, err := c.Client.KSQL.List(context.Background(), req)
	if err != nil {
		return err
	}

	if isApp {
		_, _ = fmt.Fprintln(os.Stderr, errors.KSQLAppDeprecateWarning)
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
