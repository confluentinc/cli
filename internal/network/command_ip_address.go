package network

import "github.com/spf13/cobra"

func (c *command) newIpAddressCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip-address",
		Short: "List Confluent Cloud egress public IP addresses.",
	}

	cmd.AddCommand(c.newIpAddressListCommand())

	return cmd
}
