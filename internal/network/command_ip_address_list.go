package network

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type listIpAddressHumanOut struct {
	AddressType string `human:"Address Type"`
	IpPrefix    string `human:"IP Prefix"`
	Cloud       string `human:"Cloud"`
	Region      string `human:"Region"`
	Services    string `human:"Services"`
}
type listIpAddressSerializedOut struct {
	AddressType string   `serialized:"address_type"`
	IpPrefix    string   `serialized:"ip_prefix"`
	Cloud       string   `serialized:"cloud"`
	Region      string   `serialized:"region"`
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

	list := output.NewList(cmd)
	for _, ipAddress := range ipAddresses {
		if output.GetFormat(cmd) == output.Human {
			list.Add(&listIpAddressHumanOut{
				AddressType: ipAddress.GetAddressType(),
				IpPrefix:    ipAddress.GetIpPrefix(),
				Cloud:       ipAddress.GetCloud(),
				Region:      ipAddress.GetRegion(),
				Services:    strings.Join(ipAddress.GetServices().Items, ", "),
			})
		} else {
			list.Add(&listIpAddressSerializedOut{
				AddressType: ipAddress.GetAddressType(),
				IpPrefix:    ipAddress.GetIpPrefix(),
				Cloud:       ipAddress.GetCloud(),
				Region:      ipAddress.GetRegion(),
				Services:    ipAddress.GetServices().Items,
			})
		}
	}
	return list.Print()
}
