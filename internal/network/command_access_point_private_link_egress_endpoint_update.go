package network

import (
	"github.com/spf13/cobra"

	networkingoutboundprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-outbound-privatelink/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *accessPointCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing egress endpoint.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of egress endpoint "ap-123456".`,
				Code: "confluent network access-point private-link egress-endpoint update ap-123456 --name my-new-egress-endpoint",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the egress endpoint.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagRequired("name")

	return cmd
}

func (c *accessPointCommand) update(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	updateAccessPoint := networkingoutboundprivatelinkv1.NetworkingV1AccessPointUpdate{
		Spec: &networkingoutboundprivatelinkv1.NetworkingV1AccessPointSpecUpdate{
			DisplayName: networkingoutboundprivatelinkv1.PtrString(name),
			Environment: &networkingoutboundprivatelinkv1.ObjectReference{Id: environmentId},
		},
	}

	accessPoint, err := c.V2Client.UpdateAccessPoint(args[0], updateAccessPoint)
	if err != nil {
		return err
	}

	return printPrivateLinkEgressEndpointTable(cmd, accessPoint)
}
