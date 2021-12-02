package connect

import (
	"context"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	connectLogListFields = []string{"ClusterId", "EnvironmentId", "ServiceAccountId", "TopicName"}
	humanLabelMap        = map[string]string{
		"ClusterId":        "Cluster",
		"EnvironmentId":    "Environment",
		"ServiceAccountId": "Service Account",
		"TopicName":        "Topic Name",
	}
	structuredLabelMap = map[string]string{
		"ClusterId":        "cluster_id",
		"EnvironmentId":    "environment_id",
		"ServiceAccountId": "service_account_id",
		"TopicName":        "topic_name",
	}
)

type connectLogEventsInfo struct {
	ClusterId        string
	EnvironmentId    string
	ServiceAccountId string
	TopicName        string
}

func (c *eventCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe the Connect log events configuration.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.describe),
	}

	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)

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

	info := &connectLogEventsInfo{
		ClusterId:        auditLog.ClusterId,
		EnvironmentId:    auditLog.AccountId,
		ServiceAccountId: serviceAccount.ResourceId,
		TopicName:        "confluent-connect-log-events",
	}

	return output.DescribeObject(cmd, info, connectLogListFields, humanLabelMap, structuredLabelMap)
}
