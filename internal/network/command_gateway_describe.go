package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newGatewayDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a gateway.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validGatewayArgs),
		RunE:              c.gatewayDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe gateway "gw-123456".`,
				Code: "confluent network gateway describe gw-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) gatewayDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	gateway, err := c.V2Client.GetGateway(environmentId, args[0])
	if err != nil {
		return err
	}

	return printGatewayTable(cmd, gateway)
}
