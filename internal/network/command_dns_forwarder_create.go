package network

import (
	"github.com/confluentinc/cli/v4/pkg/properties"
	"github.com/spf13/cobra"
	"os"
	"strings"

	networkingdnsforwarderv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking-dnsforwarder/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

const (
	forwardViaIp  = "ForwardViaIp"
	forwardViaGCP = "ForwardViaGcpDnsZones"
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
				Code: "confluent network dns forwarder create --dns-server-ips 10.200.0.0,10.201.0.0 --gateway gw-123456 --domains abc.com,def.com ",
			},
			examples.Example{
				Text: "Create a named DNS forwarder.",
				Code: "confluent network dns forwarder create my-dns-forwarder --dns-server-ips 10.200.0.0,10.201.0.0 --gateway gw-123456 --domains abc.com,def.com ",
			},
		),
	}

	addGatewayFlag(cmd, c.AuthenticatedCLICommand)
	addForwarderFlags(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("dns-server-ips", "domain-mapping")
	cobra.CheckErr(cmd.MarkFlagRequired("gateway"))
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

	var createDnsForwarder networkingdnsforwarderv1.NetworkingV1DnsForwarder
	if len(dnsServerIps) != 0 {
		createDnsForwarder = networkingdnsforwarderv1.NetworkingV1DnsForwarder{
			Spec: &networkingdnsforwarderv1.NetworkingV1DnsForwarderSpec{
				Domains: &domains,
				Config: &networkingdnsforwarderv1.NetworkingV1DnsForwarderSpecConfigOneOf{
					NetworkingV1ForwardViaIp: &networkingdnsforwarderv1.NetworkingV1ForwardViaIp{
						Kind:         forwardViaIp,
						DnsServerIps: dnsServerIps,
					},
				},
				Environment: &networkingdnsforwarderv1.ObjectReference{Id: environmentId},
				Gateway:     &networkingdnsforwarderv1.ObjectReference{Id: gateway},
			},
		}
	} else {
		domain, err := cmd.Flags().GetString("domain-mapping")
		if err != nil {
			return err
		}
		domainMap, err := DomainFlagToMap(domain)
		if err != nil {
			return err
		}
		createDnsForwarder = networkingdnsforwarderv1.NetworkingV1DnsForwarder{
			Spec: &networkingdnsforwarderv1.NetworkingV1DnsForwarderSpec{
				Domains: &domains,
				Config: &networkingdnsforwarderv1.NetworkingV1DnsForwarderSpecConfigOneOf{
					NetworkingV1ForwardViaGcpDnsZones: &networkingdnsforwarderv1.NetworkingV1ForwardViaGcpDnsZones{
						Kind:           forwardViaGCP,
						DomainMappings: domainMap,
					},
				},
				Environment: &networkingdnsforwarderv1.ObjectReference{Id: environmentId},
				Gateway:     &networkingdnsforwarderv1.ObjectReference{Id: gateway},
			},
		}
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

func DomainFlagToMap(domains string) (map[string]networkingdnsforwarderv1.NetworkingV1ForwardViaGcpDnsZonesDomainMappings, error) {
	buf, err := os.ReadFile(domains)
	if err != nil {
		return nil, err
	}

	domainsContent := properties.ParseLines(string(buf))
	m := make(map[string]networkingdnsforwarderv1.NetworkingV1ForwardViaGcpDnsZonesDomainMappings)
	for i := len(domainsContent) - 1; i >= 0; i-- {
		x := strings.SplitN(domainsContent[i], "=", 2)
		if _, ok := m[x[0]]; !ok {
			y := strings.SplitN(x[1], ",", 2)
			zone := replaceSpecialCharacters(y[0])
			project := replaceSpecialCharacters(y[1])
			m[x[0]] = networkingdnsforwarderv1.NetworkingV1ForwardViaGcpDnsZonesDomainMappings{Zone: networkingdnsforwarderv1.PtrString(zone), Project: networkingdnsforwarderv1.PtrString(project)}
		}

	}
	return m, nil
}

func replaceSpecialCharacters(val string) string {
	// Replace \\n, \\r and \\t with newline, carriage return and tab characters as specified in
	// https://docs.oracle.com/cd/E23095_01/Platform.93/ATGProgGuide/html/s0204propertiesfileformat01.html.
	return strings.ReplaceAll(strings.ReplaceAll(
		strings.ReplaceAll(val, "\\n", "\n"), "\\r", "\r"), "\\t", "\t")
}
