package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *accessPointCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List egress endpoints.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *accessPointCommand) list(cmd *cobra.Command, _ []string) error {
	accessPoints, err := c.getEgressEndpoints()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, accessPoint := range accessPoints {
		if accessPoint.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if accessPoint.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		out := &accessPointOut{
			Id:          accessPoint.GetId(),
			Name:        accessPoint.Spec.GetDisplayName(),
			Gateway:     accessPoint.Spec.Gateway.GetId(),
			Environment: accessPoint.Spec.Environment.GetId(),
			Phase:       accessPoint.Status.GetPhase(),
		}

		if accessPoint.Status.Config != nil && accessPoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus != nil {
			out.AwsVpcEndpointService = accessPoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus.GetVpcEndpointId()
		}

		if accessPoint.Status.Config != nil && accessPoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus != nil {
			out.AzurePrivateLinkPrivateEndpoint = accessPoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointResourceId()
		}

		list.Add(out)
	}

	return list.Print()
}
