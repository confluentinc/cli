package network

import (
	"fmt"
	"slices"
	"time"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcloud "github.com/confluentinc/cli/v4/pkg/cloud"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/network"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

type out struct {
	Id                                         string            `human:"ID" serialized:"id"`
	Environment                                string            `human:"Environment" serialized:"environment"`
	Name                                       string            `human:"Name,omitempty" serialized:"name,omitempty"`
	Gateway                                    string            `human:"Gateway,omitempty" serialized:"gateway,omitempty"`
	Cloud                                      string            `human:"Cloud" serialized:"cloud"`
	Region                                     string            `human:"Region" serialized:"region"`
	Cidr                                       string            `human:"CIDR,omitempty" serialized:"cidr,omitempty"`
	Zones                                      []string          `human:"Zones,omitempty" serialized:"zones,omitempty"`
	Phase                                      string            `human:"Phase" serialized:"phase"`
	SupportedConnectionTypes                   []string          `human:"Supported Connection Types" serialized:"supported_connection_types"`
	ActiveConnectionTypes                      []string          `human:"Active Connection Types,omitempty" serialized:"active_connection_types,omitempty"`
	AwsVpc                                     string            `human:"AWS VPC,omitempty" serialized:"aws_vpc,omitempty"`
	AwsAccount                                 string            `human:"AWS Account,omitempty" serialized:"aws_account,omitempty"`
	AwsPrivateLinkEndpointService              string            `human:"AWS Private Link Endpoint Service,omitempty" serialized:"aws_private_link_endpoint_service,omitempty"`
	GcpProject                                 string            `human:"GCP Project,omitempty" serialized:"gcp_project,omitempty"`
	GcpVpcNetwork                              string            `human:"GCP VPC Network,omitempty" serialized:"gcp_vpc_network,omitempty"`
	GcpPrivateServiceConnectServiceAttachments map[string]string `human:"GCP Private Service Connect Service Attachments,omitempty" serialized:"gcp_private_service_connect_service_attachments,omitempty"`
	AzureVNet                                  string            `human:"Azure VNet,omitempty" serialized:"azure_vnet,omitempty"`
	AzureSubscription                          string            `human:"Azure Subscription,omitempty" serialized:"azure_subscription,omitempty"`
	AzurePrivateLinkServiceAliases             map[string]string `human:"Azure Private Link Service Aliases,omitempty" serialized:"azure_private_link_service_aliases,omitempty"`
	AzurePrivateLinkServiceResourceIds         map[string]string `human:"Azure Private Link Service Resources,omitempty" serialized:"azure_private_link_service_resource_ids,omitempty"`
	DnsResolution                              string            `human:"DNS Resolution,omitempty" serialized:"dns_resolution,omitempty"`
	DnsDomain                                  string            `human:"DNS Domain,omitempty" serialized:"dns_domain,omitempty"`
	ZonalSubdomains                            map[string]string `human:"Zonal Subdomains,omitempty" serialized:"zonal_subdomains,omitempty"`
	ZoneInfo                                   []string          `human:"Zone Info,omitempty" serialized:"zone_info,omitempty"`
	IdleSince                                  time.Time         `human:"Idle Since,omitempty" serialized:"idle_since,omitempty"`
}

type command struct {
	*pcmd.AuthenticatedCLICommand
}

var (
	ConnectionTypes = []string{"privatelink", "peering", "transitgateway"}
	DnsResolutions  = []string{"private", "chased-private"}
)

func New(prerunner pcmd.PreRunner, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "network",
		Short:       "Manage Confluent Cloud networks.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(newAccessPointCommand(prerunner, cfg))
	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newDnsCommand())
	cmd.AddCommand(c.newGatewayCommand())
	cmd.AddCommand(c.newIpAddressCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newNetworkLinkCommand())
	cmd.AddCommand(c.newPeeringCommand())
	cmd.AddCommand(c.newPrivateLinkCommand())
	cmd.AddCommand(c.newRegionCommand())
	cmd.AddCommand(c.newTransitGatewayAttachmentCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printTable(cmd *cobra.Command, network networkingv1.NetworkingV1Network) error {
	if network.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if network.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	cloud := network.Spec.GetCloud()
	phase := network.Status.GetPhase()
	supportedConnectionTypes := network.Status.GetSupportedConnectionTypes()

	zoneInfoStr, err := formatZoneInfoItems(network.Spec.GetZonesInfo())
	if err != nil {
		return err
	}

	human := &out{
		Id:                       network.GetId(),
		Environment:              network.Spec.Environment.GetId(),
		Name:                     network.Spec.GetDisplayName(),
		Gateway:                  network.Spec.GetGateway().Id,
		Cloud:                    cloud,
		Region:                   network.Spec.GetRegion(),
		Zones:                    network.Spec.GetZones(),
		ZoneInfo:                 zoneInfoStr,
		Phase:                    phase,
		SupportedConnectionTypes: supportedConnectionTypes,
		ActiveConnectionTypes:    network.Status.GetActiveConnectionTypes(),
	}

	describeFields := []string{"Id", "Environment", "Name", "Gateway", "Cloud", "Region", "Zones", "Phase", "SupportedConnectionTypes", "ActiveConnectionTypes", "ZoneInfo"}

	if slices.Contains(supportedConnectionTypes, "PRIVATELINK") {
		human.DnsResolution = network.Spec.DnsConfig.GetResolution()
		human.DnsDomain = network.Status.GetDnsDomain()
		human.ZonalSubdomains = network.Status.GetZonalSubdomains()
		describeFields = append(describeFields, "DnsResolution", "DnsDomain", "ZonalSubdomains")

		if network.Status.Cloud == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "cloud")
		}

		switch cloud {
		case pcloud.Aws:
			human.AwsPrivateLinkEndpointService = network.Status.Cloud.NetworkingV1AwsNetwork.GetPrivateLinkEndpointService()
			describeFields = append(describeFields, "AwsPrivateLinkEndpointService")
		case pcloud.Gcp:
			human.GcpPrivateServiceConnectServiceAttachments = network.Status.Cloud.NetworkingV1GcpNetwork.GetPrivateServiceConnectServiceAttachments()
			describeFields = append(describeFields, "GcpPrivateServiceConnectServiceAttachments")
		case pcloud.Azure:
			human.AzurePrivateLinkServiceAliases = network.Status.Cloud.NetworkingV1AzureNetwork.GetPrivateLinkServiceAliases()
			human.AzurePrivateLinkServiceResourceIds = network.Status.Cloud.NetworkingV1AzureNetwork.GetPrivateLinkServiceResourceIds()
			describeFields = append(describeFields, "AzurePrivateLinkServiceAliases", "AzurePrivateLinkServiceResourceIds")
		}
	} else {
		human.Cidr = network.Spec.GetCidr()
		describeFields = append(describeFields, "Cidr")
	}

	if phase == "READY" {
		if network.Status.Cloud == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "cloud")
		}

		switch cloud {
		case pcloud.Aws:
			human.AwsVpc = network.Status.Cloud.NetworkingV1AwsNetwork.GetVpc()
			human.AwsAccount = network.Status.Cloud.NetworkingV1AwsNetwork.GetAccount()
			describeFields = append(describeFields, "AwsVpc", "AwsAccount")
		case pcloud.Gcp:
			human.GcpVpcNetwork = network.Status.Cloud.NetworkingV1GcpNetwork.GetVpcNetwork()
			human.GcpProject = network.Status.Cloud.NetworkingV1GcpNetwork.GetProject()
			describeFields = append(describeFields, "GcpVpcNetwork", "GcpProject")
		case pcloud.Azure:
			human.AzureVNet = network.Status.Cloud.NetworkingV1AzureNetwork.GetVnet()
			human.AzureSubscription = network.Status.Cloud.NetworkingV1AzureNetwork.GetSubscription()
			describeFields = append(describeFields, "AzureVNet", "AzureSubscription")
		}
	}

	if !network.Status.GetIdleSince().IsZero() {
		human.IdleSince = network.Status.GetIdleSince()
		describeFields = append(describeFields, "IdleSince")
	}

	table := output.NewTable(cmd)
	table.Add(human)
	table.Filter(describeFields)
	return table.Print()
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	return c.validArgsMultiple(cmd, args)
}

func (c *command) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	return autocompleteNetworks(c.V2Client, environmentId)
}

func autocompleteNetworks(client *ccloudv2.Client, environmentId string) []string {
	networks, err := getNetworks(client, environmentId, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(networks))
	for i, network := range networks {
		suggestions[i] = fmt.Sprintf("%s\t%s", network.GetId(), network.Spec.GetDisplayName())
	}
	return suggestions
}

func getNetworks(client *ccloudv2.Client, environmentId string, name, cloud, region, cidr, phase, connectionType []string) ([]networkingv1.NetworkingV1Network, error) {
	return client.ListNetworks(environmentId, name, cloud, region, cidr, phase, connectionType)
}

func addConnectionTypesFlag(cmd *cobra.Command) {
	cmd.Flags().StringSlice("connection-types", nil, fmt.Sprintf(`A comma-separated list of network access types: %s.`, utils.ArrayToCommaDelimitedString(ConnectionTypes, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "connection-types", func(_ *cobra.Command, _ []string) []string { return ConnectionTypes })
}

func addDnsResolutionFlag(cmd *cobra.Command) {
	cmd.Flags().String("dns-resolution", "", fmt.Sprintf("Specify the DNS resolution as %s.", utils.ArrayToCommaDelimitedString(DnsResolutions, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "dns-resolution", func(_ *cobra.Command, _ []string) []string { return DnsResolutions })
}

func addNetworkFlag(cmd *cobra.Command, c *pcmd.AuthenticatedCLICommand) {
	cmd.Flags().String("network", "", "Network ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "network", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return nil
		}

		return autocompleteNetworks(c.V2Client, environmentId)
	})
}

func addListNetworkFlag(cmd *cobra.Command, c *pcmd.AuthenticatedCLICommand) {
	cmd.Flags().StringSlice("network", nil, "A comma-separated list of network IDs.")
	pcmd.RegisterFlagCompletionFunc(cmd, "network", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return nil
		}

		return autocompleteNetworks(c.V2Client, environmentId)
	})
}

func (c *command) addPrivateLinkAttachmentFlag(cmd *cobra.Command) {
	cmd.Flags().String("attachment", "", "Private link attachment ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "attachment", c.validPrivateLinkAttachmentArgsMultiple)
}

func addAcceptedNetworksFlag(cmd *cobra.Command, c *pcmd.AuthenticatedCLICommand) {
	cmd.Flags().StringSlice("accepted-networks", nil, "A comma-separated list of networks from which connections can be accepted.")
	pcmd.RegisterFlagCompletionFunc(cmd, "accepted-networks", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return nil
		}

		return autocompleteNetworks(c.V2Client, environmentId)
	})
}

func addAcceptedEnvironmentsFlag(cmd *cobra.Command, command *pcmd.AuthenticatedCLICommand) {
	cmd.Flags().StringSlice("accepted-environments", nil, "A comma-separated list of environments from which connections can be accepted.")
	pcmd.RegisterFlagCompletionFunc(cmd, "accepted-environments", func(cmd *cobra.Command, args []string) []string {
		if err := command.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return pcmd.AutocompleteEnvironments(command.Client, command.V2Client)
	})
}

func (c *command) addNetworkLinkServiceFlag(cmd *cobra.Command) {
	cmd.Flags().String("network-link-service", "", "Network link service ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "network-link-service", c.validNetworkLinkServicesArgsMultiple)
}

func (c *command) addListNetworkLinkServiceFlag(cmd *cobra.Command) {
	cmd.Flags().StringSlice("network-link-service", nil, "A comma-separated list of network link service IDs.")
	pcmd.RegisterFlagCompletionFunc(cmd, "network-link-service", c.validNetworkLinkServicesArgsMultiple)
}

func (c *command) addRegionFlagNetwork(cmd *cobra.Command, command *pcmd.AuthenticatedCLICommand) {
	cmd.Flags().String("region", "", "Cloud region ID for this network.")
	pcmd.RegisterFlagCompletionFunc(cmd, "region", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		cloud, _ := cmd.Flags().GetString("cloud")
		regions, err := network.ListRegions(command.Client, cloud)
		if err != nil {
			return nil
		}

		suggestions := make([]string, len(regions))
		for i, region := range regions {
			suggestions[i] = region.RegionId
		}
		return suggestions
	})
}

func (c *command) addListRegionFlagNetwork(cmd *cobra.Command, command *pcmd.AuthenticatedCLICommand) {
	cmd.Flags().StringSlice("region", nil, "A comma-separated list of cloud region IDs.")
	pcmd.RegisterFlagCompletionFunc(cmd, "region", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		cloud, _ := cmd.Flags().GetString("cloud")
		regions, err := network.ListRegions(command.Client, cloud)
		if err != nil {
			return nil
		}

		suggestions := make([]string, len(regions))
		for i, region := range regions {
			suggestions[i] = region.RegionId
		}
		return suggestions
	})
}

func addPhaseFlag(cmd *cobra.Command, resourceType string) {
	cmd.Flags().StringSlice("phase", nil, "A comma-separated list of phases.")
	pcmd.RegisterFlagCompletionFunc(cmd, "phase", func(_ *cobra.Command, _ []string) []string {
		switch resourceType {
		case resource.NetworkLinkService:
			return []string{"ready"}
		case resource.NetworkLinkEndpoint:
			return []string{"provisioning", "pending-accept", "ready", "failed", "deprovisioning", "expired", "disconnected", "disconnecting", "inactive"}
		case resource.PrivateLinkAccess:
			return []string{"provisioning", "ready", "failed", "deprovisioning"}
		case resource.Peering:
			return []string{"provisioning", "pending-accept", "ready", "failed", "deprovisioning", "disconnected"}
		case resource.PrivateLinkAttachment:
			return []string{"provisioning", "waiting-for-connections", "ready", "failed", "expired", "deprovisioning"}
		case resource.TransitGatewayAttachment:
			return []string{"provisioning", "ready", "pending-accept", "failed", "deprovisioning", "disconnected", "error"}
		case resource.Network:
			return []string{"provisioning", "ready", "failed", "deprovisioning"}
		case resource.NetworkLinkServiceAssociation:
			return []string{"provisioning", "pending-accept", "ready", "failed", "deprovisioning", "expired", "disconnected", "disconnecting", "inactive"}
		default:
			return nil
		}
	})
}

func toUpper(strSlice []string) []string {
	for i, str := range strSlice {
		strSlice[i] = ccloudv2.ToUpper(str)
	}
	return strSlice
}

func addForwarderFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("domains", nil, "A comma-separated list of domains for the DNS forwarder to use.")
	cmd.Flags().StringSlice("dns-server-ips", nil, "A comma-separated list of IP addresses for the DNS server.")
	cmd.Flags().String("domain-mapping", "", "Path to a domain mapping file containing domain mappings. Each mapping should have the format of domain=zone,project. Mappings are separated by new-line characters.")
}

func addGatewayFlag(cmd *cobra.Command, c *pcmd.AuthenticatedCLICommand) {
	cmd.Flags().String("gateway", "", "Gateway ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "gateway", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return nil
		}

		return autocompleteGateways(c.V2Client, environmentId)
	})
}

func formatZoneInfoItems(zoneInfoItems []networkingv1.NetworkingV1ZoneInfo) ([]string, error) {
	var formattedItems []string

	for _, item := range zoneInfoItems {
		if item.ZoneId != nil && item.Cidr != nil {
			formattedItems = append(formattedItems, fmt.Sprintf("%s=%s", *item.ZoneId, *item.Cidr))
		} else if item.Cidr != nil {
			formattedItems = append(formattedItems, *item.Cidr)
		} else {
			return []string{}, fmt.Errorf("zone info item is missing CIDR")
		}
	}

	return formattedItems, nil
}
