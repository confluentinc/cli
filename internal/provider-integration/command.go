package providerintegration

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

type providerIntegrationOut struct {
	Id                 string   `human:"ID" serialized:"id"`
	Name               string   `human:"Name" serialized:"name"`
	Provider           string   `human:"Provider" serialized:"provider"`
	Environment        string   `human:"Environment" serialized:"environment"`
	IamRoleArn         string   `human:"IAM Role ARN" serialized:"iam_role_arn"`
	ExternalId         string   `human:"External ID" serialized:"external_id"`
	CustomerIamRoleArn string   `human:"Customer IAM Role ARN" serialized:"customer_iam_role_arn"`
	Usages             []string `human:"Usages" serialized:"usages"`
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "provider-integration",
		Aliases:     []string{"pi"},
		Short:       "Manage CLI provider integrations.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}
