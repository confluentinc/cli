package network

import (
	"strings"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <name>",
		Short:   "Create a new network.",
		Args:    cobra.ExactArgs(1),
		RunE:    c.create,
		Example: examples.BuildExampleString(examples.Example{}),
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("region", "", "Cloud region ID for this network.")
	pcmd.AddConnectionTypesFlag(cmd)
	cmd.Flags().String("cidr", "", `Specify a /16 IPv4 CIDR block to be used. Required for networks of connection type "peering" and "transitgateway".`)
	cmd.Flags().StringSlice("zones", nil, `Specify the availability zones for this network seperating with commas (e.g. "use1-az1,use1-az2,use1-az3").`)
	cmd.Flags().StringSlice("zone-info", nil, `Specify a comma-separated list of "zone=cidr" pairs. Each CIDR must be a /27 IPv4 CIDR block.`)
	pcmd.AddDnsResolutionFlag(cmd)
	cmd.Flags().String("reserved-cidr", "", `Specify a /24 IPv4 CIDR block to be used. Can be used for AWS networks of connection type "peering" and "transitgateway".`)

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))
	cobra.CheckErr(cmd.MarkFlagRequired("connection-types"))

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	name := args[0]

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

	connectionTypesItems := make([]string, len(connectionTypes))
	for i, ct := range connectionTypes {
		connectionTypesItems[i] = strings.ToUpper(ct)
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
			DisplayName:     networkingv1.PtrString(name),
			Cloud:           networkingv1.PtrString(cloud),
			Region:          networkingv1.PtrString(region),
			ConnectionTypes: &networkingv1.NetworkingV1ConnectionTypes{Items: connectionTypesItems},
			Environment:     &networkingv1.ObjectReference{Id: environmentId},
		},
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
		zoneInfo := strings.Split(info, "=")
		if len(zoneInfo) != 2 {
			return nil, errors.NewErrorWithSuggestions("invalid zones-info", "zone1=cidr1") // TO-DO
		}
		zoneInfoItems[i] = networkingv1.NetworkingV1ZoneInfo{
			ZoneId: networkingv1.PtrString(zoneInfo[0]), Cidr: networkingv1.PtrString(zoneInfo[1]),
		}
	}
	return zoneInfoItems, nil
}
