package providerintegration

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a provider integration.",
		Long:  "Describe a provider integration, specified by the given provider integration ID.\n\n⚠️  DEPRECATION NOTICE: This command will be deprecated in Q4 2025. Use 'confluent provider-integration v2 describe' for new integrations.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe provider integration "cspi-12345" in the current environment.`,
				Code: "confluent provider-integration describe cspi-12345",
			},
			examples.Example{
				Text: `Describe provider integration "cspi-12345" in environment "env-abcdef".`,
				Code: "confluent provider-integration describe cspi-12345 --environment env-abcdef",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	id := args[0]
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	providerIntegration, err := c.V2Client.DescribeProviderIntegration(id, environmentId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	resp := providerIntegrationOut{
		Id:              providerIntegration.GetId(),
		Name:            providerIntegration.GetDisplayName(),
		Provider:        providerIntegration.GetProvider(),
		Environment:     providerIntegration.Environment.GetId(),
		IamRoleArn:      providerIntegration.Config.PimV1AwsIntegrationConfig.GetIamRoleArn(),
		ExternalId:      providerIntegration.Config.PimV1AwsIntegrationConfig.GetExternalId(),
		CustomerRoleArn: providerIntegration.Config.PimV1AwsIntegrationConfig.GetCustomerIamRoleArn(),
		Usages:          providerIntegration.GetUsages(),
	}

	table.Add(&resp)
	return table.Print()
}
