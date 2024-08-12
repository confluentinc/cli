package network

import (
	"sort"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type listIpAddressOut struct {
	Cloud       string   `human:"Cloud" serialized:"cloud"`
	Region      string   `human:"Region" serialized:"region"`
	IpPrefix    string   `human:"IP Prefix" serialized:"ip_prefix"`
	AddressType string   `human:"Address Type" serialized:"address_type"`
	Services    []string `human:"Services" serialized:"services"`
}

func (c *command) newIpAddressListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluent Cloud egress public IP addresses.",
		Args:  cobra.NoArgs,
		RunE:  c.ipAddressList,
	}

	pcmd.AddListCloudFlag(cmd)
	c.addListRegionFlagNetwork(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().StringSlice("services", nil, "A comma-separated list of services.")
	cmd.Flags().StringSlice("address-type", nil, "A comma-separated list of address-types.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) ipAddressList(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetStringSlice("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetStringSlice("region")
	if err != nil {
		return err
	}

	services, err := cmd.Flags().GetStringSlice("services")
	if err != nil {
		return err
	}

	addressType, err := cmd.Flags().GetStringSlice("address-type")
	if err != nil {
		return err
	}

	cloud, services, addressType = toUpper(cloud), toUpper(services), toUpper(addressType)

	ipAddresses, err := c.V2Client.ListIpAddresses(cloud, region, services, addressType)
	if err != nil {
		return err
	}

	// Sort ipAddresses by Cloud then Region then IpPrefix ASC.
	sort.Slice(ipAddresses, func(i, j int) bool {
		cloudI := ipAddresses[i].GetCloud()
		cloudJ := ipAddresses[j].GetCloud()

		if cloudI == cloudJ {
			regionI := ipAddresses[i].GetRegion()
			regionJ := ipAddresses[j].GetRegion()

			if regionI == regionJ {
				return ipAddresses[i].GetIpPrefix() < ipAddresses[j].GetIpPrefix()
			}
			return regionI < regionJ
		}

		return cloudI < cloudJ
	})

	list := output.NewList(cmd)
	for _, ipAddress := range ipAddresses {
		list.Add(&listIpAddressOut{
			IpPrefix:    ipAddress.GetIpPrefix(),
			Cloud:       ipAddress.GetCloud(),
			Region:      ipAddress.GetRegion(),
			AddressType: ipAddress.GetAddressType(),
			Services:    ipAddress.GetServices().Items,
		})
	}

	// Disable default sort to use the custom sort above.
	list.Sort(false)
	return list.Print()
}
