package providerintegration

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more provider integrations.",
		Long:  "Delete one or more provider integrations, specified by the given provider integration ID.\n\n⚠️  DEPRECATION NOTICE: This command will be deprecated in Q4 2025. Use 'confluent provider-integration v2 delete' for new integrations.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete the provider integration "cspi-12345" in the current environment.`,
				Code: "confluent provider-integration delete cspi-12345",
			},
			examples.Example{
				Text: `Delete the provider integrations "cspi-12345" and "cspi-67890" in environment "env-abcdef".`,
				Code: "confluent provider-integration delete cspi-12345 cspi-67890 --environment env-abcdef",
			},
		),
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddOutputFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.DescribeProviderIntegration(id, environmentId)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.ProviderIntegration); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteProviderIntegration(id, environmentId)
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.ProviderIntegration)
	return err
}
