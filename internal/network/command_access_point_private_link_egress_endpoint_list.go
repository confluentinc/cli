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
	egressEndpoints, err := c.getEgressEndpoints()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, egressEndpoint := range egressEndpoints {
		if egressEndpoint.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if egressEndpoint.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		out := &accessPointOut{
			Id:          egressEndpoint.GetId(),
			Name:        egressEndpoint.Spec.GetDisplayName(),
			Gateway:     egressEndpoint.Spec.Gateway.GetId(),
			Environment: egressEndpoint.Spec.Environment.GetId(),
			Phase:       egressEndpoint.Status.GetPhase(),
		}

		if egressEndpoint.Status.Config != nil && egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus != nil {
			out.AwsVpcEndpointService = egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus.GetVpcEndpointId()
		}

		if egressEndpoint.Status.Config != nil && egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus != nil {
			out.AzurePrivateLinkPrivateEndpoint = egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointResourceId()
		}

		list.Add(out)
	}

	return list.Print()
}
