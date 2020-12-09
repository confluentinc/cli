package auditlog

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

var (
	listFields    = []string{"ClusterId", "EnvironmentId", "ServiceAccountId", "TopicName"}
	humanLabelMap = map[string]string{
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

type describeCmd struct {
	*pcmd.AuthenticatedCLICommand
}

type auditLogStruct struct {
	ClusterId        string
	EnvironmentId    string
	ServiceAccountId int32
	TopicName        string
}

func NewDescribeCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := &describeCmd{
		pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:   "describe",
				Short: "Describe the audit log configuration for this organization.",
				Args:  cobra.NoArgs,
			},
			prerunner,
		),
	}
	c.RunE = pcmd.NewCLIRunE(c.describe)
	c.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	c.Flags().SortFlags = false
	return c.Command
}

func (c describeCmd) describe(cmd *cobra.Command, _ []string) error {
	if c.State.Auth == nil || c.State.Auth.Organization == nil || c.State.Auth.Organization.AuditLog == nil {
		return errors.New(errors.AuditLogsNotEnabledErrorMsg)
	}
	auditLog := c.State.Auth.Organization.AuditLog
	return output.DescribeObject(cmd, &auditLogStruct{
		ClusterId:        auditLog.ClusterId,
		EnvironmentId:    auditLog.AccountId,
		ServiceAccountId: auditLog.ServiceAccountId,
		TopicName:        auditLog.TopicName,
	}, listFields, humanLabelMap, structuredLabelMap)
}
