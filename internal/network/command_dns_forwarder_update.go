package network

import (
	"github.com/spf13/cobra"

	networkingdnsforwarderv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-dnsforwarder/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newDnsForwarderUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing DNS forwarder.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validDnsForwarderArgs),
		RunE:              c.dnsForwarderUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of DNS forwarder "dnsf-123456".`,
				Code: "confluent network dns forwarder update dnsf-123456 --name my-new-dns-forwarder",
			},
			examples.Example{
				Text: `Update the DNS server IPs and domains of DNS forwarder "dnsf-123456".`,
				Code: "confluent network dns forwarder update dnsf-123456 --dns-server-ips 10.200.0.0,10.201.0.0 --domains abc.com,def.com",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the DNS forwarder.")
	addForwarderFlags(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)
	cmd.MarkFlagsOneRequired("name", "dns-server-ips", "domain-mapping", "domains")

	return cmd
}

func (c *command) dnsForwarderUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	dnsForwarder, err := c.V2Client.GetDnsForwarder(environmentId, args[0])
	if err != nil {
		return err
	}

	updateDnsForwarder := networkingdnsforwarderv1.NetworkingV1DnsForwarderUpdate{
		Spec: &networkingdnsforwarderv1.NetworkingV1DnsForwarderSpecUpdate{
			Environment: &networkingdnsforwarderv1.ObjectReference{Id: environmentId},
		},
	}
	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		updateDnsForwarder.Spec.SetDisplayName(name)
	}

	if cmd.Flags().Changed("domains") {
		domains, err := cmd.Flags().GetStringSlice("domains")
		if err != nil {
			return err
		}
		updateDnsForwarder.Spec.SetDomains(domains)
	}

	if cmd.Flags().Changed("dns-server-ips") {
		updateDnsForwarder.Spec.Config = &networkingdnsforwarderv1.NetworkingV1DnsForwarderSpecUpdateConfigOneOf{NetworkingV1ForwardViaIp: dnsForwarder.Spec.Config.NetworkingV1ForwardViaIp}
		dnsServerIps, err := cmd.Flags().GetStringSlice("dns-server-ips")
		if err != nil {
			return err
		}
		updateDnsForwarder.Spec.Config.NetworkingV1ForwardViaIp.SetDnsServerIps(dnsServerIps)
	} else if cmd.Flags().Changed("domain-mapping") {
		updateDnsForwarder.Spec.Config = &networkingdnsforwarderv1.NetworkingV1DnsForwarderSpecUpdateConfigOneOf{NetworkingV1ForwardViaGcpDnsZones: dnsForwarder.Spec.Config.NetworkingV1ForwardViaGcpDnsZones}
		domain, err := cmd.Flags().GetString("domain-mapping")
		if err != nil {
			return err
		}
		domainMap, err := DomainFlagToMap(domain)
		if err != nil {
			return err
		}
		updateDnsForwarder.Spec.Config.NetworkingV1ForwardViaGcpDnsZones.SetDomainMappings(domainMap)
	}

	forwarder, err := c.V2Client.UpdateDnsForwarder(environmentId, args[0], updateDnsForwarder)
	if err != nil {
		return err
	}

	return printDnsForwarderTable(cmd, forwarder)
}
