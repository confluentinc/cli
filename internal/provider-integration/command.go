package providerintegration

import (
	"github.com/spf13/cobra"

	v2 "github.com/confluentinc/cli/v4/internal/provider-integration/v2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

type providerIntegrationOut struct {
	Id              string   `human:"ID" serialized:"id"`
	Name            string   `human:"Name" serialized:"name"`
	Provider        string   `human:"Provider" serialized:"provider"`
	Environment     string   `human:"Environment" serialized:"environment"`
	IamRoleArn      string   `human:"IAM Role ARN" serialized:"iam_role_arn"`
	ExternalId      string   `human:"External ID" serialized:"external_id"`
	CustomerRoleArn string   `human:"Customer Role ARN" serialized:"customer_role_arn"`
	Usages          []string `human:"Usages" serialized:"usages"`
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "provider-integration",
		Aliases:     []string{"pi"},
		Short:       "Manage Confluent Cloud provider integrations.",
		Long:        "Manage Confluent Cloud provider integrations.\n\n⚠️  DEPRECATION NOTICE: Provider Integration v1 commands support AWS only and will be deprecated in CLI v5. Provider Integration v2 currently supports GCP and Azure, and in CLI v5 will support all clouds (AWS, GCP, Azure). Use 'confluent provider-integration v2' for new integrations.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(v2.New(prerunner))

	return cmd
}
