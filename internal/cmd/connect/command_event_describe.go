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

	serviceAccount, err := c.PrivateClient.User.GetServiceAccount(context.Background(), auditLog.GetServiceAccountId())
	if err != nil {
		return err
	}

	info := &connectLogEventsInfo{
		ClusterId:        auditLog.GetClusterId(),
		EnvironmentId:    auditLog.GetAccountId(),
		ServiceAccountId: serviceAccount.GetResourceId(),
		TopicName:        "confluent-connect-log-events",
	}

	return output.DescribeObject(cmd, info, connectLogListFields, humanLabelMap, structuredLabelMap)
}
