package network

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

type humanOut struct {
	Id                                         string    `human:"ID"`
	EnvironmentId                              string    `human:"Environment"`
	Name                                       string    `human:"Name,omitempty"`
	Gateway                                    string    `human:"Gateway,omitempty"`
	Cloud                                      string    `human:"Cloud"`
	Region                                     string    `human:"Region"`
	Cidr                                       string    `human:"CIDR,omitempty"`
	Zones                                      string    `human:"Zones,omitempty"`
	Phase                                      string    `human:"Phase"`
	SupportedConnectionTypes                   string    `human:"Supported Connection Types"`
	ActiveConnectionTypes                      string    `human:"Active Connection Types,omitempty"`
	AwsVpc                                     string    `human:"AWS VPC,omitempty"`
	AwsAccount                                 string    `human:"AWS Account,omitempty"`
	AwsPrivateLinkEndpointService              string    `human:"AWS Private Link Endpoint Service,omitempty"`
	GcpProject                                 string    `human:"GCP Project,omitempty"`
	GcpVpcNetwork                              string    `human:"GCP VPC Network,omitempty"`
	GcpPrivateServiceConnectServiceAttachments string    `human:"GCP Private Service Connect Service Attachments,omitempty"`
	AzureVNet                                  string    `human:"Azure VNet,omitempty"`
	AzureSubscription                          string    `human:"Azure Subscription,omitempty"`
	AzurePrivateLinkServiceAliases             string    `human:"Azure Private Link Service Aliases,omitempty"`
	AzurePrivateLinkServiceResourceIds         string    `human:"Azure Private Link Service Resources,omitempty"`
	DnsResolution                              string    `human:"DNS Resolution,omitempty"`
	DnsDomain                                  string    `human:"DNS Domain,omitempty"`
	ZonalSubdomains                            string    `human:"Zonal Subdomains,omitempty"`
	IdleSince                                  time.Time `human:"Idle Since,omitempty"`
}

type serializedOut struct {
	Id                                         string            `serialized:"id"`
	EnvironmentId                              string            `serialized:"environment_id"`
	Name                                       string            `serialized:"name,omitempty"`
	Gateway                                    string            `serialized:"gateway,omitempty"`
	Cloud                                      string            `serialized:"cloud"`
	Region                                     string            `serialized:"region"`
	Cidr                                       string            `serialized:"cidr,omitempty"`
	Zones                                      []string          `serialized:"zones,omitempty"`
	Phase                                      string            `serialized:"phase"`
	SupportedConnectionTypes                   []string          `serialized:"supported_connection_types"`
	ActiveConnectionTypes                      []string          `serialized:"active_connection_types,omitempty"`
	AwsVpc                                     string            `serialized:"aws_vpc,omitempty"`
	AwsAccount                                 string            `serialized:"aws_account,omitempty"`
	AwsPrivateLinkEndpointService              string            `serialized:"aws_private_link_endpoint_service,omitempty"`
	GcpProject                                 string            `serialized:"gcp_project,omitempty"`
	GcpVpcNetwork                              string            `serialized:"gcp_vpc_network,omitempty"`
	GcpPrivateServiceConnectServiceAttachments map[string]string `serialized:"gcp_private_service_connect_service_attachments,omitempty"`
	AzureVNet                                  string            `serialized:"azure_vnet,omitempty"`
	AzureSubscription                          string            `serialized:"azure_subscription,omitempty"`
	AzurePrivateLinkServiceAliases             map[string]string `serialized:"azure_private_link_service_aliases,omitempty"`
	AzurePrivateLinkServiceResourceIds         map[string]string `serialized:"azure_private_link_service_resource_ids,omitempty"`
	DnsResolution                              string            `serialized:"dns_resolution,omitempty"`
	DnsDomain                                  string            `serialized:"dns_domain,omitempty"`
	ZonalSubdomains                            map[string]string `serialized:"zonal_subdomains,omitempty"`
	IdleSince                                  time.Time         `serialized:"idle_since,omitempty"`
}

type command struct {
	*pcmd.AuthenticatedCLICommand
}

const (
	CloudAws   = "AWS"
	CloudAzure = "AZURE"
	CloudGcp   = "GCP"
)

var (
	ConnectionTypes     = []string{"privatelink", "peering", "transitgateway"}
	DnsResolutions      = []string{"private", "chased-private"}
	DnsForwarderConfigs = []string{"forward-via-ip"}
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "network",
		Short:       "Manage Confluent Cloud networks.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newDnsCommand())
	cmd.AddCommand(c.newIpAddressCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newPeeringCommand())
	cmd.AddCommand(c.newPrivateLinkCommand())
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

	if output.GetFormat(cmd) == output.Human {
		return printHumanTable(cmd, network)
	}

	return printSerializedTable(cmd, network)
}

func printHumanTable(cmd *cobra.Command, network networkingv1.NetworkingV1Network) error {
	cloud := network.Spec.GetCloud()
	phase := network.Status.GetPhase()
	supportedConnectionTypes := network.Status.GetSupportedConnectionTypes().Items

	human := &humanOut{
		Id:                       network.GetId(),
		EnvironmentId:            network.Spec.Environment.GetId(),
		Name:                     network.Spec.GetDisplayName(),
		Gateway:                  network.Spec.GetGateway().Id,
		Cloud:                    cloud,
		Region:                   network.Spec.GetRegion(),
		Zones:                    strings.Join(network.Spec.GetZones(), ", "),
		Phase:                    phase,
		SupportedConnectionTypes: strings.Join(supportedConnectionTypes, ", "),
		ActiveConnectionTypes:    strings.Join(network.Status.GetActiveConnectionTypes().Items, ", "),
	}

	describeFields := []string{"Id", "EnvironmentId", "Name", "Gateway", "Cloud", "Region", "Zones", "Phase", "SupportedConnectionTypes", "ActiveConnectionTypes"}

	if slices.Contains(supportedConnectionTypes, "PRIVATELINK") {
		human.DnsResolution = network.Spec.DnsConfig.GetResolution()
		human.DnsDomain = network.Status.GetDnsDomain()
		human.ZonalSubdomains = convertMapToString(network.Status.GetZonalSubdomains())
		describeFields = append(describeFields, "DnsResolution", "DnsDomain", "ZonalSubdomains")

		if network.Status.Cloud == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "cloud")
		}

		switch cloud {
		case CloudAws:
			human.AwsPrivateLinkEndpointService = network.Status.Cloud.NetworkingV1AwsNetwork.GetPrivateLinkEndpointService()
			describeFields = append(describeFields, "AwsPrivateLinkEndpointService")
		case CloudGcp:
			human.GcpPrivateServiceConnectServiceAttachments = convertMapToString(network.Status.Cloud.NetworkingV1GcpNetwork.GetPrivateServiceConnectServiceAttachments())
			describeFields = append(describeFields, "GcpPrivateServiceConnectServiceAttachments")
		case CloudAzure:
			human.AzurePrivateLinkServiceAliases = convertMapToString(network.Status.Cloud.NetworkingV1AzureNetwork.GetPrivateLinkServiceAliases())
			human.AzurePrivateLinkServiceResourceIds = convertMapToString(network.Status.Cloud.NetworkingV1AzureNetwork.GetPrivateLinkServiceResourceIds())
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
		case CloudAws:
			human.AwsVpc = network.Status.Cloud.NetworkingV1AwsNetwork.GetVpc()
			human.AwsAccount = network.Status.Cloud.NetworkingV1AwsNetwork.GetAccount()
			describeFields = append(describeFields, "AwsVpc", "AwsAccount")
		case CloudGcp:
			human.GcpVpcNetwork = network.Status.Cloud.NetworkingV1GcpNetwork.GetVpcNetwork()
			human.GcpProject = network.Status.Cloud.NetworkingV1GcpNetwork.GetProject()
			describeFields = append(describeFields, "GcpVpcNetwork", "GcpProject")
		case CloudAzure:
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

func printSerializedTable(cmd *cobra.Command, network networkingv1.NetworkingV1Network) error {
	cloud := network.Spec.GetCloud()
	phase := network.Status.GetPhase()
	supportedConnectionTypes := network.Status.GetSupportedConnectionTypes().Items

	serialized := &serializedOut{
		Id:                       network.GetId(),
		EnvironmentId:            network.Spec.Environment.GetId(),
		Name:                     network.Spec.GetDisplayName(),
		Gateway:                  network.Spec.GetGateway().Id,
		Cloud:                    network.Spec.GetCloud(),
		Region:                   network.Spec.GetRegion(),
		Zones:                    network.Spec.GetZones(),
		Phase:                    network.Status.GetPhase(),
		SupportedConnectionTypes: supportedConnectionTypes,
		ActiveConnectionTypes:    network.Status.GetActiveConnectionTypes().Items,
	}

	describeFields := []string{"Id", "EnvironmentId", "Name", "Gateway", "Cloud", "Region", "Zones", "Phase", "SupportedConnectionTypes", "ActiveConnectionTypes"}

	if slices.Contains(supportedConnectionTypes, "PRIVATELINK") {
		serialized.DnsResolution = network.Spec.DnsConfig.GetResolution()
		serialized.DnsDomain = network.Status.GetDnsDomain()
		serialized.ZonalSubdomains = network.Status.GetZonalSubdomains()
		describeFields = append(describeFields, "DnsResolution", "DnsDomain", "ZonalSubdomains")

		if network.Status.Cloud == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "cloud")
		}

		switch cloud {
		case CloudAws:
			serialized.AwsPrivateLinkEndpointService = network.Status.Cloud.NetworkingV1AwsNetwork.GetPrivateLinkEndpointService()
			describeFields = append(describeFields, "AwsPrivateLinkEndpointService")
		case CloudGcp:
			serialized.GcpPrivateServiceConnectServiceAttachments = network.Status.Cloud.NetworkingV1GcpNetwork.GetPrivateServiceConnectServiceAttachments()
			describeFields = append(describeFields, "GcpPrivateServiceConnectServiceAttachments")
		case CloudAzure:
			serialized.AzurePrivateLinkServiceAliases = network.Status.Cloud.NetworkingV1AzureNetwork.GetPrivateLinkServiceAliases()
			serialized.AzurePrivateLinkServiceResourceIds = network.Status.Cloud.NetworkingV1AzureNetwork.GetPrivateLinkServiceResourceIds()
			describeFields = append(describeFields, "AzurePrivateLinkServiceAliases", "AzurePrivateLinkServiceResourceIds")
		}
	} else {
		serialized.Cidr = network.Spec.GetCidr()
		describeFields = append(describeFields, "Cidr")
	}

	if phase == "READY" {
		if network.Status.Cloud == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "cloud")
		}

		switch cloud {
		case CloudAws:
			serialized.AwsVpc = network.Status.Cloud.NetworkingV1AwsNetwork.GetVpc()
			serialized.AwsAccount = network.Status.Cloud.NetworkingV1AwsNetwork.GetAccount()
			describeFields = append(describeFields, "AwsVpc", "AwsAccount")
		case CloudGcp:
			serialized.GcpVpcNetwork = network.Status.Cloud.NetworkingV1GcpNetwork.GetVpcNetwork()
			serialized.GcpProject = network.Status.Cloud.NetworkingV1GcpNetwork.GetProject()
			describeFields = append(describeFields, "GcpVpcNetwork", "GcpProject")
		case CloudAzure:
			serialized.AzureVNet = network.Status.Cloud.NetworkingV1AzureNetwork.GetVnet()
			serialized.AzureSubscription = network.Status.Cloud.NetworkingV1AzureNetwork.GetSubscription()
			describeFields = append(describeFields, "AzureVNet", "AzureSubscription")
		}
	}

	if !network.Status.GetIdleSince().IsZero() {
		serialized.IdleSince = network.Status.GetIdleSince()
		describeFields = append(describeFields, "IdleSince")
	}

	table := output.NewTable(cmd)
	table.Add(serialized)

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
	networks, err := getNetworks(client, environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(networks))
	for i, network := range networks {
		suggestions[i] = fmt.Sprintf("%s\t%s", network.GetId(), network.Spec.GetDisplayName())
	}
	return suggestions
}

func getNetworks(client *ccloudv2.Client, environmentId string) ([]networkingv1.NetworkingV1Network, error) {
	return client.ListNetworks(environmentId)
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

func (c *command) addPrivateLinkAttachmentFlag(cmd *cobra.Command) {
	cmd.Flags().String("attachment", "", "Private link attachment ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "attachment", c.validPrivateLinkAttachmentArgsMultiple)
}

func convertMapToString(m map[string]string) string {
	items := make([]string, len(m))

	i := 0
	for key, val := range m {
		items[i] = fmt.Sprintf("%s=%s", key, val)
		i++
	}

	sort.Strings(items)
	return strings.Join(items, ", ")
}

func addConfigFlag(cmd *cobra.Command) {
	cmd.Flags().String("config", "", fmt.Sprintf("Specify the DNS forwarder configuration as %s.", utils.ArrayToCommaDelimitedString(DnsForwarderConfigs, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "config", func(_ *cobra.Command, _ []string) []string { return DnsForwarderConfigs })
}

func addForwarderFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("dns-server-ips", nil, "A comma-separated list of IP addresses for the DNS server.")
	cmd.Flags().StringSlice("domains", nil, "A comma-separated list of domains for the DNS forwarder to use.")
}
