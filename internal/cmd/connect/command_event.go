package connect

import (
	"context"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type eventCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type connectLogEventsInfo struct {
	ClusterId        string
	EnvironmentId    string
	ServiceAccountId string
	TopicName        string
}

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

// NewEventCommand returns the Cobra command for Connect log.
func NewEventCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &eventCommand{
		pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:         "event",
				Short:       "Manage Connect log events configuration.",
				Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
			},
			prerunner,
		),
	}
	cmd.init()
	return cmd.Command
}

func (c *eventCommand) init() {
	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe the Connect log events configuration.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.describe),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	c.AddCommand(describeCmd)
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

	return output.DescribeObject(cmd, &connectLogEventsInfo{
		ClusterId:        auditLog.ClusterId,
		EnvironmentId:    auditLog.AccountId,
		ServiceAccountId: serviceAccount.ResourceId,
		TopicName:        "confluent-connect-log-events",
	}, connectLogListFields, humanLabelMap, structuredLabelMap)
}
