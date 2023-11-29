package network

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type listIpAddressHumanOut struct {
	Cloud       string `human:"Cloud"`
	Region      string `human:"Region"`
	IpPrefix    string `human:"IP Prefix"`
	AddressType string `human:"Address Type"`
	Services    string `human:"Services"`
}

type listIpAddressSerializedOut struct {
	Cloud       string   `serialized:"cloud"`
	Region      string   `serialized:"region"`
	IpPrefix    string   `serialized:"ip_prefix"`
	AddressType string   `serialized:"address_type"`
	Services    []string `serialized:"services"`
}

func (c *command) newIpAddressListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluent Cloud egress public IP addresses.",
		Args:  cobra.NoArgs,
		RunE:  c.ipAddressList,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) ipAddressList(cmd *cobra.Command, _ []string) error {
	ipAddresses, err := c.V2Client.ListIpAddresses()
	if err != nil {
		return err
	}

	// Sort ipAddresses by Cloud then Region then IpPrefix ASC.
	sort.Slice(ipAddresses, func(i, j int) bool {
		ipAddressI := ipAddresses[i]
		ipAddressJ := ipAddresses[j]
		cloudI := ipAddressI.GetCloud()
		cloudJ := ipAddressJ.GetCloud()
		regionI := ipAddressI.GetRegion()
		regionJ := ipAddressJ.GetRegion()

		if cloudI == cloudJ {
			if regionI == regionJ {
				return ipAddressI.GetIpPrefix() < ipAddressJ.GetIpPrefix()
			}
			return regionI < regionJ
		}

		return cloudI < cloudJ
	})

	list := output.NewList(cmd)
	for _, ipAddress := range ipAddresses {
		if output.GetFormat(cmd) == output.Human {
			list.Add(&listIpAddressHumanOut{
				IpPrefix:    ipAddress.GetIpPrefix(),
				Cloud:       ipAddress.GetCloud(),
				Region:      ipAddress.GetRegion(),
				AddressType: ipAddress.GetAddressType(),
				Services:    strings.Join(ipAddress.GetServices().Items, ", "),
			})
		} else {
			list.Add(&listIpAddressSerializedOut{
				IpPrefix:    ipAddress.GetIpPrefix(),
				Cloud:       ipAddress.GetCloud(),
				Region:      ipAddress.GetRegion(),
				AddressType: ipAddress.GetAddressType(),
				Services:    ipAddress.GetServices().Items,
			})
		}
	}

	// Disable default sort to use the custom sort above.
	list.Sort(false)
	return list.Print()
}
