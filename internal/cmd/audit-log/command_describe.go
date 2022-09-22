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
	ClusterId        string `human:"Cluster" json:"cluster_id" yaml:"cluster_id"`
	EnvironmentId    string `human:"Environment" json:"environment_id" yaml:"environment_id"`
	ServiceAccountId string `human:"Service Account" json:"service_account_id" yaml:"service_account_id"`
	TopicName        string `human:"Topic Name" json:"topic_name" yaml:"topic_name"`
}

func newDescribeCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe",
		Short:       "Describe the audit log configuration for this organization.",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &describeCmd{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	c.RunE = c.describe

	pcmd.AddOutputFlag(c.Command)

	return c.Command
}

func (c describeCmd) describe(cmd *cobra.Command, _ []string) error {
	if auditLog := v1.GetAuditLog(c.State); auditLog == nil {
		return errors.New(errors.AuditLogsNotEnabledErrorMsg)
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
		TopicName:        auditLog.TopicName,
	})
	return table.Print()
}
