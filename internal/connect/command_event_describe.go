package connect

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type eventDescribeOut struct {
	Cluster        string `human:"Cluster" serialized:"cluster"`
	Environment    string `human:"Environment" serialized:"environment"`
	ServiceAccount string `human:"Service Account" serialized:"service_account"`
	TopicName      string `human:"Topic Name" serialized:"topic_name"`
}

func (c *eventCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe the Connect log events configuration.",
		Args:  cobra.NoArgs,
		RunE:  c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *eventCommand) describe(cmd *cobra.Command, _ []string) error {
	auditLog := c.Context.GetOrganization().GetAuditLog()

	if auditLog.GetClusterId() == "" {
		return fmt.Errorf("Connect Log Events are not enabled for this organization")
	}

	table := output.NewTable(cmd)
	table.Add(&eventDescribeOut{
		Cluster:        auditLog.GetClusterId(),
		Environment:    auditLog.GetAccountId(),
		ServiceAccount: auditLog.GetServiceAccountResourceId(),
		TopicName:      "confluent-connect-log-events",
	})
	return table.Print()
}
