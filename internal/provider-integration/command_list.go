package providerintegration

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List provider integrations.",
		Long:  "List provider integrations, optionally filtered by cloud provider.\n\n⚠️  DEPRECATION NOTICE: Provider Integration v1 supports AWS only and will be deprecated in CLI v5. Provider Integration v2 currently supports GCP and Azure, and in CLI v5 will support all clouds (AWS, GCP, Azure). Use 'confluent provider-integration v2 list' for new integrations.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List provider integrations in the current environment.",
				Code: "confluent provider-integration list",
			},
			examples.Example{
				Text: `List provider integrations in environment "env-abcdef".`,
				Code: "confluent provider-integration list --environment env-abcdef",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cloud, _ := cmd.Flags().GetString("cloud")

	providerIntegrations, err := c.V2Client.ListProviderIntegrations(cloud, environmentId)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)

	for _, providerIntegration := range providerIntegrations {
		list.Add(&providerIntegrationOut{
			Id:              providerIntegration.GetId(),
			Name:            providerIntegration.GetDisplayName(),
			Provider:        providerIntegration.GetProvider(),
			Environment:     providerIntegration.Environment.GetId(),
			IamRoleArn:      providerIntegration.Config.PimV1AwsIntegrationConfig.GetIamRoleArn(),
			ExternalId:      providerIntegration.Config.PimV1AwsIntegrationConfig.GetExternalId(),
			CustomerRoleArn: providerIntegration.Config.PimV1AwsIntegrationConfig.GetCustomerIamRoleArn(),
			Usages:          providerIntegration.GetUsages(),
		})
	}

	return list.Print()
}
