package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *accessPointCommand) newIngressEndpointListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ingress endpoints.",
		Args:  cobra.NoArgs,
		RunE:  c.listIngressEndpoint,
	}

	cmd.Flags().StringSlice("names", nil, "A comma-separated list of display names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *accessPointCommand) listIngressEndpoint(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	names, err := cmd.Flags().GetStringSlice("names")
	if err != nil {
		return err
	}

	ingressEndpoints, err := c.V2Client.ListNetworkAccessPoints(environmentId, names)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, ingressEndpoint := range ingressEndpoints {
		if ingressEndpoint.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if ingressEndpoint.Spec.GetConfig().NetworkingV1AwsIngressPrivateLinkEndpoint == nil &&
			ingressEndpoint.Spec.GetConfig().NetworkingV1AzureIngressPrivateLinkEndpoint == nil &&
			ingressEndpoint.Spec.GetConfig().NetworkingV1GcpIngressPrivateServiceConnectEndpoint == nil {
			continue
		}
		if ingressEndpoint.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		out := &ingressEndpointOut{
			Id:          ingressEndpoint.GetId(),
			Name:        ingressEndpoint.Spec.GetDisplayName(),
			Gateway:     ingressEndpoint.Spec.Gateway.GetId(),
			Environment: ingressEndpoint.Spec.Environment.GetId(),
			Phase:       ingressEndpoint.Status.GetPhase(),
		}

		if ingressEndpoint.Status.Config != nil && ingressEndpoint.Status.Config.NetworkingV1AwsIngressPrivateLinkEndpointStatus != nil {
			out.AwsVpcEndpointId = ingressEndpoint.Status.Config.NetworkingV1AwsIngressPrivateLinkEndpointStatus.GetVpcEndpointId()
			out.AwsVpcEndpointServiceName = ingressEndpoint.Status.Config.NetworkingV1AwsIngressPrivateLinkEndpointStatus.GetVpcEndpointServiceName()
			if ingressEndpoint.Status.Config.NetworkingV1AwsIngressPrivateLinkEndpointStatus.HasDnsDomain() {
				out.DnsDomain = ingressEndpoint.Status.Config.NetworkingV1AwsIngressPrivateLinkEndpointStatus.GetDnsDomain()
			}
		}

		if ingressEndpoint.Status.Config != nil && ingressEndpoint.Status.Config.NetworkingV1AzureIngressPrivateLinkEndpointStatus != nil {
			out.AzurePrivateLinkServiceAlias = ingressEndpoint.Status.Config.NetworkingV1AzureIngressPrivateLinkEndpointStatus.GetPrivateLinkServiceAlias()
			out.AzurePrivateLinkServiceResourceId = ingressEndpoint.Status.Config.NetworkingV1AzureIngressPrivateLinkEndpointStatus.GetPrivateLinkServiceResourceId()
			out.AzurePrivateEndpointResourceId = ingressEndpoint.Status.Config.NetworkingV1AzureIngressPrivateLinkEndpointStatus.GetPrivateEndpointResourceId()
			if ingressEndpoint.Status.Config.NetworkingV1AzureIngressPrivateLinkEndpointStatus.HasDnsDomain() {
				out.DnsDomain = ingressEndpoint.Status.Config.NetworkingV1AzureIngressPrivateLinkEndpointStatus.GetDnsDomain()
			}
		}

		if ingressEndpoint.Status.Config != nil && ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus != nil {
			out.GcpPrivateServiceConnectServiceAttachment = ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus.GetPrivateServiceConnectServiceAttachment()
			out.GcpPrivateServiceConnectConnectionId = ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus.GetPrivateServiceConnectConnectionId()
			if ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus.HasDnsDomain() {
				out.DnsDomain = ingressEndpoint.Status.Config.NetworkingV1GcpIngressPrivateServiceConnectEndpointStatus.GetDnsDomain()
			}
		}

		list.Add(out)
	}

	return list.Print()
}
