package network

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type egressEndpointOut struct {
	Id                                         string   `human:"ID" serialized:"id"`
	Name                                       string   `human:"Name,omitempty" serialized:"name,omitempty"`
	Environment                                string   `human:"Environment" serialized:"environment"`
	Gateway                                    string   `human:"Gateway" serialized:"gateway"`
	Phase                                      string   `human:"Phase" serialized:"phase"`
	AwsVpcEndpointService                      string   `human:"AWS VPC Endpoint Service,omitempty" serialized:"aws_vpc_endpoint_service,omitempty"`
	AwsVpcEndpoint                             string   `human:"AWS VPC Endpoint,omitempty" serialized:"aws_vpc_endpoint,omitempty"`
	AwsVpcEndpointDnsName                      string   `human:"AWS VPC Endpoint DNS Name,omitempty" serialized:"aws_vpc_endpoint_dns_name,omitempty"`
	AzurePrivateLinkService                    string   `human:"Azure Private Link Service,omitempty" serialized:"azure_private_link_service,omitempty"`
	AzurePrivateLinkSubresourceName            string   `human:"Azure Private Link Subresource Name,omitempty" serialized:"azure_private_link_subresource_name,omitempty"`
	AzurePrivateEndpoint                       string   `human:"Azure Private Endpoint,omitempty" serialized:"azure_private_endpoint,omitempty"`
	AzurePrivateEndpointDomain                 string   `human:"Azure Private Endpoint Domain,omitempty" serialized:"azure_private_endpoint_domain,omitempty"`
	AzurePrivateEndpointIpAddress              string   `human:"Azure Private Endpoint IP Address,omitempty" serialized:"azure_private_endpoint_ip_address,omitempty"`
	AzurePrivateEndpointCustomDnsConfigDomains []string `human:"Azure Private Endpoint Custom DNS Config Domains,omitempty" serialized:"azure_private_endpoint_custom_dns_config_domains,omitempty"`
	HighAvailability                           bool     `human:"High Availability,omitempty" serialized:"high_availability,omitempty"`
}

func (c *accessPointCommand) newEgressEndpointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "egress-endpoint",
		Short: "Manage private link egress endpoints.",
	}

	cmd.AddCommand(c.newEgressEndpointCreateCommand())
	cmd.AddCommand(c.newEgressEndpointDeleteCommand())
	cmd.AddCommand(c.newEgressEndpointDescribeCommand())
	cmd.AddCommand(c.newEgressEndpointListCommand())
	cmd.AddCommand(c.newEgressEndpointUpdateCommand())

	return cmd
}

func (c *accessPointCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *accessPointCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteEgressEndpoints()
}

func (c *accessPointCommand) autocompleteEgressEndpoints() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	accessPoints, err := c.V2Client.ListAccessPoints(environmentId, nil)
	if err != nil {
		return nil
	}
	egressEndpoints := slices.DeleteFunc(accessPoints, func(accessPoint networkingaccesspointv1.NetworkingV1AccessPoint) bool {
		return accessPoint.Spec.GetConfig().NetworkingV1AwsEgressPrivateLinkEndpoint == nil && accessPoint.Spec.GetConfig().NetworkingV1AzureEgressPrivateLinkEndpoint == nil
	})

	suggestions := make([]string, len(egressEndpoints))
	for i, egressEndpoint := range egressEndpoints {
		suggestions[i] = fmt.Sprintf("%s\t%s", egressEndpoint.GetId(), egressEndpoint.Spec.GetDisplayName())
	}
	return suggestions
}

func printPrivateLinkEgressEndpointTable(cmd *cobra.Command, egressEndpoint networkingaccesspointv1.NetworkingV1AccessPoint) error {
	if egressEndpoint.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
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
	}
	if egressEndpoint.Spec.Config != nil && egressEndpoint.Spec.Config.NetworkingV1AzureEgressPrivateLinkEndpoint != nil {
		out.AzurePrivateLinkService = egressEndpoint.Spec.Config.NetworkingV1AzureEgressPrivateLinkEndpoint.GetPrivateLinkServiceResourceId()
		out.AzurePrivateLinkSubresourceName = egressEndpoint.Spec.Config.NetworkingV1AzureEgressPrivateLinkEndpoint.GetPrivateLinkSubresourceName()
	}

	if egressEndpoint.Status.Config != nil && egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus != nil {
		out.AwsVpcEndpoint = egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus.GetVpcEndpointId()
		out.AwsVpcEndpointDnsName = egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus.GetVpcEndpointDnsName()
	}
	if egressEndpoint.Status.Config != nil && egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus != nil {
		out.AzurePrivateEndpoint = egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointResourceId()
		out.AzurePrivateEndpointDomain = egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointDomain()
		out.AzurePrivateEndpointIpAddress = egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointIpAddress()
		out.AzurePrivateEndpointCustomDnsConfigDomains = egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointCustomDnsConfigDomains()
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.PrintWithAutoWrap(false)
}
