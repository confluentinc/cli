package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a network.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a Confluent network in AWS with connection type "transitgateway" by specifying zones and CIDR.`,
				Code: "confluent network create --cloud aws --region us-west-2 --connection-types transitgateway --zones usw2-az1,usw2-az2,usw2-az4 --cidr 10.1.0.0/16",
			},
			examples.Example{
				Text: `Create a named Confluent network in AWS with connection type "transitgateway" by specifying zones and CIDR.`,
				Code: "confluent network create aws-tgw-network --cloud aws --region us-west-2 --connection-types transitgateway --zones usw2-az1,usw2-az2,usw2-az4 --cidr 10.1.0.0/16",
			},
			examples.Example{
				Text: `Create a named Confluent network in AWS with connection types "transitgateway" and "peering" by specifying zones and CIDR.`,
				Code: "confluent network create aws-tgw-peering-network --cloud aws --region us-west-2 --connection-types transitgateway,peering --zones usw2-az1,usw2-az2,usw2-az4 --cidr 10.1.0.0/16",
			},
			examples.Example{
				Text: `Create a named Confluent network in AWS with connection type "peering" by specifying zone info.`,
				Code: "confluent network create aws-peering-network --cloud aws --region us-west-2 --connection-types peering --zone-info usw2-az1=10.10.0.0/27,usw2-az3=10.10.0.32/27,usw2-az4=10.10.0.64/27",
			},
			examples.Example{
				Text: `Create a named Confluent network in GCP with connection type "peering" by specifying zones and CIDR.`,
				Code: "confluent network create gcp-peering-network --cloud gcp --region us-central1 --connection-types peering --zones us-central1-a,us-central1-b,us-central1-c --cidr 10.1.0.0/16",
			},
			examples.Example{
				Text: `Create a named Confluent network in Azure with connection type "privatelink" by specifying DNS resolution.`,
				Code: "confluent network create azure-pl-network --cloud azure --region eastus2 --connection-types privatelink --dns-resolution chased-private",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("region", "", "Cloud region ID for this network.")
	addConnectionTypesFlag(cmd)
	cmd.Flags().String("cidr", "", `A /16 IPv4 CIDR block. Required for networks of connection type "peering" and "transitgateway".`)
	cmd.Flags().StringSlice("zones", nil, `A comma-separated list of availability zones for this network.`)
	cmd.Flags().StringSlice("zone-info", nil, `A comma-separated list of "zone=cidr" pairs or CIDR blocks. Each CIDR must be a /27 IPv4 CIDR block.`)
	addDnsResolutionFlag(cmd)
	cmd.Flags().String("reserved-cidr", "", `A /24 IPv4 CIDR block. Can be used for AWS networks of connection type "peering" and "transitgateway".`)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))
	cobra.CheckErr(cmd.MarkFlagRequired("connection-types"))

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	connectionTypes, err := cmd.Flags().GetStringSlice("connection-types")
	if err != nil {
		return err
	}

	for i, connectionType := range connectionTypes {
		connectionTypes[i] = strings.ToUpper(connectionType)
	}

	cidr, err := cmd.Flags().GetString("cidr")
	if err != nil {
		return err
	}

	zones, err := cmd.Flags().GetStringSlice("zones")
	if err != nil {
		return err
	}

	zoneInfo, err := cmd.Flags().GetStringSlice("zone-info")
	if err != nil {
		return err
	}
	zoneInfoItems, err := getZoneInfoItems(zoneInfo)
	if err != nil {
		return err
	}

	dnsResolution, err := cmd.Flags().GetString("dns-resolution")
	if err != nil {
		return err
	}
	dnsResolution = strings.ToUpper(dnsResolution)
	dnsResolution = strings.ReplaceAll(dnsResolution, "-", "_")

	reservedCidr, err := cmd.Flags().GetString("reserved-cidr")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	createNetwork := networkingv1.NetworkingV1Network{
		Spec: &networkingv1.NetworkingV1NetworkSpec{
			Cloud:           networkingv1.PtrString(cloud),
			Region:          networkingv1.PtrString(region),
			ConnectionTypes: &networkingv1.NetworkingV1ConnectionTypes{Items: connectionTypes},
			Environment:     &networkingv1.ObjectReference{Id: environmentId},
		},
	}

	if name != "" {
		createNetwork.Spec.SetDisplayName(name)
	}

	if cidr != "" {
		createNetwork.Spec.SetCidr(cidr)
	}

	if len(zones) != 0 {
		createNetwork.Spec.SetZones(zones)
	}

	if len(zoneInfoItems) != 0 {
		createNetwork.Spec.SetZonesInfo(networkingv1.NetworkingV1ZonesInfo{Items: zoneInfoItems})
	}

	if dnsResolution != "" {
		createNetwork.Spec.SetDnsConfig(networkingv1.NetworkingV1DnsConfig{Resolution: dnsResolution})
	}

	if reservedCidr != "" {
		createNetwork.Spec.SetReservedCidr(reservedCidr)
	}

	network, err := c.V2Client.CreateNetwork(createNetwork)
	if err != nil {
		return err
	}

	return printTable(cmd, network)
}

func getZoneInfoItems(zoneInfo []string) ([]networkingv1.NetworkingV1ZoneInfo, error) {
	zoneInfoItems := make([]networkingv1.NetworkingV1ZoneInfo, len(zoneInfo))
	for i, info := range zoneInfo {
		zoneInfoSplit := strings.Split(info, "=")
		if len(zoneInfoSplit) == 1 {
			zoneInfoItems[i] = networkingv1.NetworkingV1ZoneInfo{
				Cidr: networkingv1.PtrString(zoneInfoSplit[0]),
			}
		} else if len(zoneInfoSplit) == 2 {
			zoneInfoItems[i] = networkingv1.NetworkingV1ZoneInfo{
				ZoneId: networkingv1.PtrString(zoneInfoSplit[0]), Cidr: networkingv1.PtrString(zoneInfoSplit[1]),
			}
		} else {
			return nil, fmt.Errorf(`zone info "%s" is not correctly formatted as <zone-ID>=<CIDR> or <CIDR>`, info)
		}
	}
	return zoneInfoItems, nil
}
