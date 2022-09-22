package connect

import (
	"context"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type out struct {
	ClusterId        string `human:"Cluster" json:"cluster_id" yaml:"cluster_id"`
	EnvironmentId    string `human:"Environment" json:"environment_id" yaml:"environment_id"`
	ServiceAccountId string `human:"Service Account" json:"service_account_id" yaml:"service_account_id"`
	TopicName        string `human:"Topic Name" json:"topic_name" yaml:"topic_name"`
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
	if c.State.Auth == nil || c.State.Auth.Organization == nil || c.State.Auth.Organization.AuditLog == nil ||
		c.State.Auth.Organization.AuditLog.ClusterId == "" {
		return errors.New(errors.ConnectLogEventsNotEnabledErrorMsg)
	}

	auditLog := c.State.Auth.Organization.AuditLog

	serviceAccount, err := c.Client.User.GetServiceAccount(context.Background(), auditLog.ServiceAccountId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		ClusterId:        auditLog.ClusterId,
		EnvironmentId:    auditLog.AccountId,
		ServiceAccountId: serviceAccount.ResourceId,
		TopicName:        "confluent-connect-log-events",
	})
	return table.Print()
}
