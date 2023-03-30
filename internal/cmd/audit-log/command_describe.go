package auditlog

import (
	"context"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type describeCmd struct {
	*pcmd.AuthenticatedCLICommand
}

type out struct {
	ClusterId        string `human:"Cluster" serialized:"cluster_id"`
	EnvironmentId    string `human:"Environment" serialized:"environment_id"`
	ServiceAccountId string `human:"Service Account" serialized:"service_account_id"`
	TopicName        string `human:"Topic Name" serialized:"topic_name"`
}

func newDescribeCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe",
		Short:       "Describe the audit log configuration for this organization.",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &describeCmd{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.describe

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c describeCmd) describe(cmd *cobra.Command, _ []string) error {
	if v1.GetAuditLog(c.Context.Context) == nil {
		return errors.New(errors.AuditLogsNotEnabledErrorMsg)
	}

	auditLog := c.Context.GetOrganization().GetAuditLog()

	serviceAccount, err := c.Client.User.GetServiceAccount(context.Background(), auditLog.GetServiceAccountId())
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		ClusterId:        auditLog.GetClusterId(),
		EnvironmentId:    auditLog.GetAccountId(),
		ServiceAccountId: serviceAccount.GetResourceId(),
		TopicName:        auditLog.GetTopicName(),
	})
	return table.Print()
}
