package network

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	networkingdnsforwarderv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking-dnsforwarder/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type dnsForwarderOut struct {
	Id                string            `human:"ID" serialized:"id"`
	Name              string            `human:"Name,omitempty" serialized:"name,omitempty"`
	Domains           []string          `human:"Domains,omitempty" serialized:"domains,omitempty"`
	DnsServerIps      []string          `human:"DNS Server IPs,omitempty" serialized:"dns_server_ips,omitempty"`
	DnsDomainMappings map[string]string `human:"DNS Domain Mappings,omitempty" serialized:"dns_domain_mappings,omitempty"`
	Environment       string            `human:"Environment" serialized:"environment"`
	Gateway           string            `human:"Gateway" serialized:"gateway"`
	Phase             string            `human:"Phase" serialized:"phase"`
}

func (c *command) newDnsForwarderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forwarder",
		Short: "Manage DNS forwarders.",
	}

	cmd.AddCommand(c.newDnsForwarderCreateCommand())
	cmd.AddCommand(c.newDnsForwarderDeleteCommand())
	cmd.AddCommand(c.newDnsForwarderDescribeCommand())
	cmd.AddCommand(c.newDnsForwarderListCommand())
	cmd.AddCommand(c.newDnsForwarderUpdateCommand())

	return cmd
}

func (c *command) getDnsForwarders() ([]networkingdnsforwarderv1.NetworkingV1DnsForwarder, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListDnsForwarders(environmentId)
}

func (c *command) validDnsForwarderArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	return c.validDnsForwardersArgsMultiple(cmd, args)
}

func (c *command) validDnsForwardersArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteDnsForwarders()
}

func (c *command) autocompleteDnsForwarders() []string {
	forwarders, err := c.getDnsForwarders()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(forwarders))
	for i, forwarder := range forwarders {
		suggestions[i] = fmt.Sprintf("%s\t%s", forwarder.GetId(), forwarder.Spec.GetDisplayName())
	}
	return suggestions
}

func printDnsForwarderTable(cmd *cobra.Command, forwarder networkingdnsforwarderv1.NetworkingV1DnsForwarder) error {
	if forwarder.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if forwarder.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	sort.Strings(forwarder.Spec.GetDomains())
	table := output.NewTable(cmd)

	table.Add(&dnsForwarderOut{
		Id:                forwarder.GetId(),
		Name:              forwarder.Spec.GetDisplayName(),
		Domains:           forwarder.Spec.GetDomains(),
		DnsServerIps:      forwarder.Spec.Config.NetworkingV1ForwardViaIp.GetDnsServerIps(),
		DnsDomainMappings: convertToTypeString(forwarder.Spec.Config.NetworkingV1ForwardViaGcpDnsZones.GetDomainMappings()),
		Gateway:           forwarder.Spec.Gateway.GetId(),
		Environment:       forwarder.Spec.Environment.GetId(),
		Phase:             forwarder.Status.GetPhase(),
	})
	return table.Print()
}

func convertToTypeString(input map[string]networkingdnsforwarderv1.NetworkingV1ForwardViaGcpDnsZonesDomainMappings) map[string]string {
	myMap := make(map[string]string)
	for key, value := range input {
		myMap[key] = fmt.Sprintf("{%s, %s}", *value.Zone, *value.Project)
	}
	return myMap
}
