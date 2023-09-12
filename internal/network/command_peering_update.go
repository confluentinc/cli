package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newPeeringUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing peering.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPeeringArgs),
		RunE:              c.updatePeering,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of peering "peer-123456"`,
				Code: `confluent network peering update peer-123456 --name "new name"`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the peering.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *command) updatePeering(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	updatePeering := networkingv1.NetworkingV1PeeringUpdate{
		Spec: &networkingv1.NetworkingV1PeeringSpecUpdate{
			DisplayName: networkingv1.PtrString(name),
			Environment: &networkingv1.ObjectReference{Id: environmentId},
		},
	}

	peering, err := c.V2Client.UpdatePeering(environmentId, args[0], updatePeering)
	if err != nil {
		return err
	}

	return printPeeringTable(cmd, peering)
}
