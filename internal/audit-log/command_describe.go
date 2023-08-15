package auditlog

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type describeCommand struct {
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

	c := &describeCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.describe

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *describeCommand) describe(cmd *cobra.Command, _ []string) error {
	user, err := c.Client.Auth.User()
	if err != nil {
		return err
	}

	auditLog := user.GetOrganization().GetAuditLog()
	if auditLog.GetServiceAccountId() == 0 {
		return errors.New(errors.AuditLogsNotEnabledErrorMsg)
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		ClusterId:        auditLog.GetClusterId(),
		EnvironmentId:    auditLog.GetAccountId(),
		ServiceAccountId: auditLog.GetServiceAccountResourceId(),
		TopicName:        auditLog.GetTopicName(),
	})
	return table.Print()
}
