package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newNetworkLinkEndpointDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a network link endpoint.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validNetworkLinkEndpointArgs),
		RunE:              c.networkLinkEndpointDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe network link endpoint "nle-123456".`,
				Code: "confluent network network-link endpoint describe nle-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) networkLinkEndpointDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	endpoint, err := c.V2Client.GetNetworkLinkEndpoint(environmentId, args[0])
	if err != nil {
		return err
	}

	return printNetworkLinkEndpointTable(cmd, endpoint)
}
