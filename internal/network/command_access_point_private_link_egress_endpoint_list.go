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

	cmd.Flags().StringSlice("names", nil, "A comma-separated list of display names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *accessPointCommand) list(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	names, err := cmd.Flags().GetStringSlice("names")
	if err != nil {
		return err
	}

	egressEndpoints, err := c.V2Client.ListAccessPoints(environmentId, names)
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

		if egressEndpoint.Spec.Config != nil && egressEndpoint.Spec.Config.NetworkingV1AwsEgressPrivateLinkEndpoint != nil {
			out.AwsVpcEndpointService = egressEndpoint.Spec.Config.NetworkingV1AwsEgressPrivateLinkEndpoint.GetVpcEndpointServiceName()
			out.HighAvailability = egressEndpoint.Spec.Config.NetworkingV1AwsEgressPrivateLinkEndpoint.GetEnableHighAvailability()
		}

		if egressEndpoint.Status.Config != nil && egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus != nil {
			out.AwsVpcEndpoint = egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus.GetVpcEndpointId()
		}

		if egressEndpoint.Spec.Config != nil && egressEndpoint.Spec.Config.NetworkingV1AzureEgressPrivateLinkEndpoint != nil {
			out.AzurePrivateLinkService = egressEndpoint.Spec.Config.NetworkingV1AzureEgressPrivateLinkEndpoint.GetPrivateLinkServiceResourceId()
		}

		if egressEndpoint.Status.Config != nil && egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus != nil {
			out.AzurePrivateLinkEndpoint = egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointResourceId()
		}

		list.Add(out)
	}

	return list.Print()
}
