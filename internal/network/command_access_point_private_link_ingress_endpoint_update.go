package network

import (
	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *accessPointCommand) newIngressEndpointUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing ingress endpoint.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validIngressEndpointArgs),
		RunE:              c.updateIngressEndpoint,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of ingress endpoint "ap-123456".`,
				Code: "confluent network access-point private-link ingress-endpoint update ap-123456 --name my-new-ingress-endpoint",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the ingress endpoint.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *accessPointCommand) updateIngressEndpoint(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	updateIngressEndpoint := networkingaccesspointv1.NetworkingV1AccessPointUpdate{
		Spec: &networkingaccesspointv1.NetworkingV1AccessPointSpecUpdate{
			DisplayName: networkingaccesspointv1.PtrString(name),
			Environment: &networkingaccesspointv1.ObjectReference{Id: environmentId},
		},
	}

	ingressEndpoint, err := c.V2Client.UpdateAccessPoint(args[0], updateIngressEndpoint)
	if err != nil {
		return err
	}

	return printPrivateLinkIngressEndpointTable(cmd, ingressEndpoint)
}
