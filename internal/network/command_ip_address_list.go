package network

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type listIpAddressHumanOut struct {
	IpPrefix    string `human:"IP Prefix"`
	Cloud       string `human:"Cloud"`
	Region      string `human:"Region"`
	AddressType string `human:"Address Type"`
	Services    string `human:"Services"`
}

type listIpAddressSerializedOut struct {
	IpPrefix    string   `serialized:"ip_prefix"`
	Cloud       string   `serialized:"cloud"`
	Region      string   `serialized:"region"`
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

	// Sort ipAddresses by Cloud then Region ASC.
	sort.Slice(ipAddresses, func(i, j int) bool {
		cloudI := ipAddresses[i].GetCloud()
		cloudJ := ipAddresses[j].GetCloud()
		if cloudI == cloudJ {
			return ipAddresses[i].GetRegion() < ipAddresses[j].GetRegion()
		}
		return ipAddresses[i].GetCloud() < ipAddresses[j].GetCloud()
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
