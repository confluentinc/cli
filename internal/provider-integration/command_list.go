package providerintegration

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List provider integrations.",
		Long:  "List Provider Integrations, optionally filtered by cloud provider.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List Provider Integrations in current environment.",
				Code: "confluent provider-integration list",
			},
			examples.Example{
				Text: "List Provider Integrations in environment env-abcdef.",
				Code: "confluent provider-integration list --environment env-abcdef",
			},
			examples.Example{
				Text: "List Provider Integrations in current environment with AWS as cloud provider.",
				Code: "confluent provider-integration list --cloud aws",
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

	// Provider information is optional
	provider, _ := cmd.Flags().GetString("cloud")

	providerIntegrations, err := c.V2Client.ListProviderIntegrations(provider, environmentId)
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
