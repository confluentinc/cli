package providerintegration

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more provider integrations.",
		Long:  "Delete one or more Provider Integrations, specified by the given provider integration ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete the Provider Integration "cspi-12345".`,
				Code: `confluent provider-integration delete cspi-12345`,
			},
			examples.Example{
				Text: `Delete the Provider Integrations "cspi-12345" and "cspi-67890" in environment "env-abcdef".`,
				Code: `confluent provider-integration delete cspi-12345 cspi-67890 --environment env-abcdef`,
			},
		),
	}

	// Auto-complete flags
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	pcmd.AddForceFlag(cmd)

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

	_, err = deletion.Delete(args, deleteFunc, resource.ProviderIntegration)
	return err
}
