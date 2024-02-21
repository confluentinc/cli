package network

import (
	"github.com/spf13/cobra"

	networkingdnsforwarderv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-dnsforwarder/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newDnsForwarderCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a DNS forwarder.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.dnsForwarderCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a DNS forwarder.",
				Code: "confluent network dns forwarder create --dns-server-ips 10.200.0.0,10.201.0.0 --gateway gw-123456 --config ForwardViaIp --domains abc.com,def.com ",
			},
			examples.Example{
				Text: "Create a named DNS forwarder.",
				Code: "confluent network dns forwarder create my-dns-forwarder --dns-server-ips 10.200.0.0,10.201.0.0 --gateway gw-123456 --config ForwardViaIp --domains abc.com,def.com ",
			},
		),
	}

	cmd.Flags().String("gateway", "", "Gateway ID.")
	addConfigFlag(cmd)
	addForwarderFlags(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("gateway"))
	cobra.CheckErr(cmd.MarkFlagRequired("config"))
	cobra.CheckErr(cmd.MarkFlagRequired("dns-server-ips"))
	cobra.CheckErr(cmd.MarkFlagRequired("domains"))

	return cmd
}

func (c *command) dnsForwarderCreate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	domains, err := cmd.Flags().GetStringSlice("domains")
	if err != nil {
		return err
	}

	gateway, err := cmd.Flags().GetString("gateway")
	if err != nil {
		return err
	}

	dnsServerIps, err := cmd.Flags().GetStringSlice("dns-server-ips")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	config, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	createDnsForwarder := networkingdnsforwarderv1.NetworkingV1DnsForwarder{
		Spec: &networkingdnsforwarderv1.NetworkingV1DnsForwarderSpec{
			Domains: &domains,
			Config: &networkingdnsforwarderv1.NetworkingV1DnsForwarderSpecConfigOneOf{
				NetworkingV1ForwardViaIp: &networkingdnsforwarderv1.NetworkingV1ForwardViaIp{
					Kind:         config,
					DnsServerIps: dnsServerIps,
				},
			},
			Environment: &networkingdnsforwarderv1.ObjectReference{Id: environmentId},
			Gateway:     &networkingdnsforwarderv1.ObjectReference{Id: gateway},
		},
	}

	if name != "" {
		createDnsForwarder.Spec.SetDisplayName(name)
	}

	forwarder, err := c.V2Client.CreateDnsForwarder(createDnsForwarder)
	if err != nil {
		return err
	}

	return printDnsForwarderTable(cmd, forwarder)
}
