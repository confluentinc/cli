package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

type humanOut struct {
	Id                       string `human:"ID"`
	EnvironmentId            string `human:"Environment"`
	Name                     string `human:"Name"`
	Cloud                    string `human:"Cloud"`
	Region                   string `human:"Region"`
	Cidr                     string `human:"CIDR"`
	Zones                    string `human:"Zones"`
	DnsResolution            string `human:"DNS Resolution,omitempty"`
	Phase                    string `human:"Phase"`
	SupportedConnectionTypes string `human:"Supported Connection Types"`
	ActiveConnectionTypes    string `human:"Active Connection Types"`
	AwsVpc                   string `human:"AWS VPC,omitempty"`
	AwsAccount               string `human:"AWS Account,omitempty"`
	GcpProject               string `human:"GCP Project,omitempty"`
	GcpVpcNetwork            string `human:"GCP VPC Network,omitempty"`
	AzureVNet                string `human:"Azure VNet,omitempty"`
	AzureSubscription        string `human:"Azure Subscription,omitempty"`
}

type serializedOut struct {
	Id                       string   `serialized:"id"`
	EnvironmentId            string   `serialized:"environment_id"`
	Name                     string   `serialized:"name"`
	Cloud                    string   `serialized:"cloud"`
	Region                   string   `serialized:"region"`
	Cidr                     string   `serialized:"cidr"`
	Zones                    []string `serialized:"zones"`
	DnsResolution            string   `serialized:"dns_resolution"`
	Phase                    string   `serialized:"phase"`
	SupportedConnectionTypes []string `serialized:"supported_connection_types"`
	ActiveConnectionTypes    []string `serialized:"active_connection_types"`
	AwsVpc                   string   `serialized:"aws_vpc,omitempty"`
	AwsAccount               string   `serialized:"aws_account,omitempty"`
	GcpProject               string   `serialized:"gcp_project,omitempty"`
	GcpVpcNetwork            string   `serialized:"gcp_vpc_network,omitempty"`
	AzureVNet                string   `serialized:"azure_vnet,omitempty"`
	AzureSubscription        string   `serialized:"azure_subscription,omitempty"`
}

type command struct {
	*pcmd.AuthenticatedCLICommand
}

var (
	ConnectionTypes = []string{"privatelink", "peering", "transitgateway"}
	DnsResolutions  = []string{"private", "chased-private"}

	CloudAws   = "AWS"
	CloudAzure = "AZURE"
	CloudGcp   = "GCP"
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
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(newPeeringCommand(prerunner))
	cmd.AddCommand(newTransitGatewayAttachmentCommand(prerunner))
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printTable(cmd *cobra.Command, network networkingv1.NetworkingV1Network) error {
	table := output.NewTable(cmd)
	describeFields := []string{"Id", "EnvironmentId", "Name", "Cloud", "Region", "Cidr", "Zones", "DnsResolution", "Phase", "SupportedConnectionTypes", "ActiveConnectionTypes"}

	if network.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if network.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	zones := network.Spec.GetZones()
	cloud := network.Spec.GetCloud()
	phase := network.Status.GetPhase()
	supportedConnectionTypes := network.Status.GetSupportedConnectionTypes().Items
	activeConnectionTypes := network.Status.GetActiveConnectionTypes().Items

	human := &humanOut{
		Id:                       network.GetId(),
		EnvironmentId:            network.Spec.Environment.GetId(),
		Name:                     network.Spec.GetDisplayName(),
		Cloud:                    network.Spec.GetCloud(),
		Region:                   network.Spec.GetRegion(),
		Cidr:                     network.Spec.GetCidr(),
		Zones:                    strings.Join(zones, ", "),
		DnsResolution:            network.Spec.DnsConfig.GetResolution(),
		Phase:                    network.Status.GetPhase(),
		SupportedConnectionTypes: strings.Join(supportedConnectionTypes, ", "),
		ActiveConnectionTypes:    strings.Join(activeConnectionTypes, ", "),
	}

	serialized := &serializedOut{
		Id:                       network.GetId(),
		EnvironmentId:            network.Spec.Environment.GetId(),
		Name:                     network.Spec.GetDisplayName(),
		Cloud:                    network.Spec.GetCloud(),
		Region:                   network.Spec.GetRegion(),
		Cidr:                     network.Spec.GetCidr(),
		Zones:                    zones,
		DnsResolution:            network.Spec.DnsConfig.GetResolution(),
		Phase:                    network.Status.GetPhase(),
		SupportedConnectionTypes: supportedConnectionTypes,
		ActiveConnectionTypes:    activeConnectionTypes,
	}

	if phase == "READY" {
		if network.Status.Cloud == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "cloud")
		}

		switch cloud {
		case CloudAws:
			human.AwsVpc = network.Status.Cloud.NetworkingV1AwsNetwork.GetVpc()
			human.AwsAccount = network.Status.Cloud.NetworkingV1AwsNetwork.GetAccount()
			serialized.AwsVpc = network.Status.Cloud.NetworkingV1AwsNetwork.GetVpc()
			serialized.AwsAccount = network.Status.Cloud.NetworkingV1AwsNetwork.GetAccount()
			describeFields = append(describeFields, "AwsVpc", "AwsAccount")
		case CloudGcp:
			human.GcpVpcNetwork = network.Status.Cloud.NetworkingV1GcpNetwork.GetVpcNetwork()
			human.GcpProject = network.Status.Cloud.NetworkingV1GcpNetwork.GetProject()
			serialized.GcpVpcNetwork = network.Status.Cloud.NetworkingV1GcpNetwork.GetVpcNetwork()
			serialized.GcpProject = network.Status.Cloud.NetworkingV1GcpNetwork.GetProject()
			describeFields = append(describeFields, "GcpVpcNetwork", "GcpProject")
		case CloudAzure:
			human.AzureVNet = network.Status.Cloud.NetworkingV1AzureNetwork.GetVnet()
			human.AzureSubscription = network.Status.Cloud.NetworkingV1AzureNetwork.GetSubscription()
			serialized.AzureVNet = network.Status.Cloud.NetworkingV1AzureNetwork.GetVnet()
			serialized.AzureSubscription = network.Status.Cloud.NetworkingV1AzureNetwork.GetSubscription()
			describeFields = append(describeFields, "AzureVNet", "AzureSubscription")
		}
	}

	if output.GetFormat(cmd) == output.Human {
		table.Add(human)
	} else {
		table.Add(serialized)
	}

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

	return c.autocompleteNetworks()
}

func (c *command) autocompleteNetworks() []string {
	networks, err := c.getNetworks()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(networks))
	for i, network := range networks {
		suggestions[i] = fmt.Sprintf("%s\t%s", network.GetId(), network.Spec.GetDisplayName())
	}
	return suggestions
}

func (c *command) getNetworks() ([]networkingv1.NetworkingV1Network, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListNetworks(environmentId)
}

func addConnectionTypesFlag(cmd *cobra.Command) {
	cmd.Flags().StringSlice("connection-types", nil, fmt.Sprintf(`A comma-separated list of network access types: %s.`, utils.ArrayToCommaDelimitedString(ConnectionTypes, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "connection-types", func(_ *cobra.Command, _ []string) []string { return ConnectionTypes })
}

func addDnsResolutionFlag(cmd *cobra.Command) {
	cmd.Flags().String("dns-resolution", "", fmt.Sprintf("Specify the DNS resolution as %s.", utils.ArrayToCommaDelimitedString(DnsResolutions, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "dns-resolution", func(_ *cobra.Command, _ []string) []string { return DnsResolutions })
}
