package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type eventDescribeOut struct {
	ClusterId        string `human:"Cluster" serialized:"cluster_id"`
	EnvironmentId    string `human:"Environment" serialized:"environment_id"`
	ServiceAccountId string `human:"Service Account" serialized:"service_account_id"`
	TopicName        string `human:"Topic Name" serialized:"topic_name"`
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
		return errors.New(errors.ConnectLogEventsNotEnabledErrorMsg)
	}

	table := output.NewTable(cmd)
	table.Add(&eventDescribeOut{
		ClusterId:        auditLog.GetClusterId(),
		EnvironmentId:    auditLog.GetAccountId(),
		ServiceAccountId: auditLog.GetServiceAccountResourceId(),
		TopicName:        "confluent-connect-log-events",
	})
	return table.Print()
}
