package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newNetworkLinkServiceDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a network link service.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validNetworkLinkServiceArgs),
		RunE:              c.networkLinkServiceDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe network link service "nls-123456".`,
				Code: "confluent network network-link service describe nls-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) networkLinkServiceDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	service, err := c.V2Client.GetNetworkLinkService(environmentId, args[0])
	if err != nil {
		return err
	}

	return printNetworkLinkServiceTable(cmd, service)
}
