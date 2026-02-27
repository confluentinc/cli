package network

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type ingressEndpointOut struct {
	Id                        string `human:"ID" serialized:"id"`
	Name                      string `human:"Name,omitempty" serialized:"name,omitempty"`
	Environment               string `human:"Environment" serialized:"environment"`
	Gateway                   string `human:"Gateway" serialized:"gateway"`
	Phase                     string `human:"Phase" serialized:"phase"`
	AwsVpcEndpointId          string `human:"AWS VPC Endpoint ID,omitempty" serialized:"aws_vpc_endpoint_id,omitempty"`
	AwsVpcEndpointServiceName string `human:"AWS VPC Endpoint Service Name,omitempty" serialized:"aws_vpc_endpoint_service_name,omitempty"`
	DnsDomain                 string `human:"DNS Domain,omitempty" serialized:"dns_domain,omitempty"`
}

func (c *accessPointCommand) newIngressEndpointCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "ingress-endpoint",
		Short:  "Manage private link ingress endpoints.",
		Hidden: !(cfg.IsTest || featureflags.Manager.BoolVariation("networking.gateway.ingress_private_link.api.enable", cfg.Context(), config.CcloudProdLaunchDarklyClient, true, false)),
	}

	cmd.AddCommand(c.newIngressEndpointCreateCommand())
	cmd.AddCommand(c.newIngressEndpointDeleteCommand())
	cmd.AddCommand(c.newIngressEndpointDescribeCommand())
	cmd.AddCommand(c.newIngressEndpointListCommand())
	cmd.AddCommand(c.newIngressEndpointUpdateCommand())

	return cmd
}

func (c *accessPointCommand) validIngressEndpointArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validIngressEndpointArgsMultiple(cmd, args)
}

func (c *accessPointCommand) validIngressEndpointArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteIngressEndpoints()
}

func (c *accessPointCommand) autocompleteIngressEndpoints() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	accessPoints, err := c.V2Client.ListAccessPoints(environmentId, nil)
	if err != nil {
		return nil
	}
	ingressEndpoints := slices.DeleteFunc(accessPoints, func(accessPoint networkingaccesspointv1.NetworkingV1AccessPoint) bool {
		return accessPoint.Spec.GetConfig().NetworkingV1AwsIngressPrivateLinkEndpoint == nil
	})

	suggestions := make([]string, len(ingressEndpoints))
	for i, ingressEndpoint := range ingressEndpoints {
		suggestions[i] = fmt.Sprintf("%s\t%s", ingressEndpoint.GetId(), ingressEndpoint.Spec.GetDisplayName())
	}
	return suggestions
}

func printPrivateLinkIngressEndpointTable(cmd *cobra.Command, ingressEndpoint networkingaccesspointv1.NetworkingV1AccessPoint) error {
	if ingressEndpoint.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
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

	table := output.NewTable(cmd)
	table.Add(out)
	return table.PrintWithAutoWrap(false)
}
