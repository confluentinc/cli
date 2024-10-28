package network

import (
	"github.com/spf13/cobra"

	networkinggatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-gateway/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newGatewayUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a gateway.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.gatewayUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of gateway "gw-abc123".`,
				Code: "confluent network gateway update gw-abc123 --name new-name",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the gateway.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *command) gatewayUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	updateGateway := networkinggatewayv1.NetworkingV1GatewayUpdate{
		Spec: &networkinggatewayv1.NetworkingV1GatewaySpecUpdate{
			DisplayName: networkinggatewayv1.PtrString(name),
			Environment: &networkinggatewayv1.ObjectReference{Id: environmentId},
		},
	}

	gateway, err := c.V2Client.UpdateGateway(args[0], updateGateway)
	if err != nil {
		return err
	}

	return printGatewayTable(cmd, gateway)
}
