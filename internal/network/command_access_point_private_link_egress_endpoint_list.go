package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *accessPointCommand) newEgressEndpointListCommand() *cobra.Command {
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
		return err
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
		if egressEndpoint.Spec.GetConfig().NetworkingV1AwsEgressPrivateLinkEndpoint == nil && egressEndpoint.Spec.GetConfig().NetworkingV1AzureEgressPrivateLinkEndpoint == nil && egressEndpoint.Spec.GetConfig().NetworkingV1GcpEgressPrivateServiceConnectEndpoint == nil {
			continue
		}
		if egressEndpoint.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		out := &egressEndpointOut{
			Id:          egressEndpoint.GetId(),
			Name:        egressEndpoint.Spec.GetDisplayName(),
			Gateway:     egressEndpoint.Spec.Gateway.GetId(),
			Environment: egressEndpoint.Spec.Environment.GetId(),
			Phase:       egressEndpoint.Status.GetPhase(),
		}

		if egressEndpoint.Spec.Config != nil && egressEndpoint.Spec.Config.NetworkingV1AwsEgressPrivateLinkEndpoint != nil {
			out.AwsVpcEndpointService = egressEndpoint.Spec.Config.NetworkingV1AwsEgressPrivateLinkEndpoint.GetVpcEndpointServiceName()
			out.HighAvailability = egressEndpoint.Spec.Config.NetworkingV1AwsEgressPrivateLinkEndpoint.GetEnableHighAvailability()
		} else if egressEndpoint.Spec.Config != nil && egressEndpoint.Spec.Config.NetworkingV1AzureEgressPrivateLinkEndpoint != nil {
			out.AzurePrivateLinkService = egressEndpoint.Spec.Config.NetworkingV1AzureEgressPrivateLinkEndpoint.GetPrivateLinkServiceResourceId()
			out.AzurePrivateLinkSubresourceName = egressEndpoint.Spec.Config.NetworkingV1AzureEgressPrivateLinkEndpoint.GetPrivateLinkSubresourceName()
		} else if egressEndpoint.Spec.Config != nil && egressEndpoint.Spec.Config.NetworkingV1GcpEgressPrivateServiceConnectEndpoint != nil {
			out.GcpPrivateServiceConnectEndpointService = egressEndpoint.Spec.Config.NetworkingV1GcpEgressPrivateServiceConnectEndpoint.GetPrivateServiceConnectEndpointTarget()
		}

		if egressEndpoint.Status.Config != nil && egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus != nil {
			out.AwsVpcEndpoint = egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus.GetVpcEndpointId()
			out.AwsVpcEndpointDnsName = egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus.GetVpcEndpointDnsName()
		} else if egressEndpoint.Status.Config != nil && egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus != nil {
			out.AzurePrivateEndpoint = egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointResourceId()
			out.AzurePrivateEndpointDomain = egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointDomain()
			out.AzurePrivateEndpointIpAddress = egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointIpAddress()
			out.AzurePrivateEndpointCustomDnsConfigDomains = egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointCustomDnsConfigDomains()
		} else if egressEndpoint.Status.Config != nil && egressEndpoint.Status.Config.NetworkingV1GcpEgressPrivateServiceConnectEndpointStatus != nil {
			out.GcpPrivateServiceConnectEndpointConnectionId = egressEndpoint.Status.Config.NetworkingV1GcpEgressPrivateServiceConnectEndpointStatus.GetPrivateServiceConnectEndpointConnectionId()
			out.GcpPrivateServiceConnectEndpointName = egressEndpoint.Status.Config.NetworkingV1GcpEgressPrivateServiceConnectEndpointStatus.GetPrivateServiceConnectEndpointName()
			out.GcpPrivateServiceConnectEndpointIpAddress = egressEndpoint.Status.Config.NetworkingV1GcpEgressPrivateServiceConnectEndpointStatus.GetPrivateServiceConnectEndpointIpAddress()
		}

		list.Add(out)
	}

	return list.Print()
}
