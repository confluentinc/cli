package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	networkingdnsforwarderv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking-dnsforwarder/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type dnsForwarderHumanOut struct {
	Id           string `human:"ID"`
	Name         string `human:"Display Name,omitempty"`
	Domains      string `human:"Domains,omitempty"`
	DnsServerIps string `human:"DNS Server IPs"`
	Environment  string `human:"Environment"`
	Gateway      string `human:"Gateway"`
	Phase        string `human:"Phase"`
}

type dnsForwarderSerializedOut struct {
	Id           string   `serialized:"id"`
	Name         string   `serialized:"display_name,omitempty"`
	Domains      []string `serialized:"domains,omitempty"`
	DnsServerIps []string `serialized:"dns_server_ips"`
	Environment  string   `serialized:"environment"`
	Gateway      string   `serialized:"gateway"`
	Phase        string   `serialized:"phase"`
}

func (c *command) newDnsForwarderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dns-forwarder",
		Short:   "Manage DNS forwarders.",
		Aliases: []string{"dnsf"},
	}

	cmd.AddCommand(c.newDnsForwarderDeleteCommand())
	cmd.AddCommand(c.newDnsForwarderDescribeCommand())
	cmd.AddCommand(c.newDnsForwarderListCommand())

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

	table := output.NewTable(cmd)

	if output.GetFormat(cmd) == output.Human {
		table.Add(&dnsForwarderHumanOut{
			Id:           forwarder.GetId(),
			Name:         forwarder.Spec.GetDisplayName(),
			Domains:      strings.Join(forwarder.Spec.GetDomains(), ", "),
			DnsServerIps: strings.Join(forwarder.Spec.Config.NetworkingV1ForwardViaIp.GetDnsServerIps(), ", "),
			Gateway:      forwarder.Spec.Gateway.GetId(),
			Environment:  forwarder.Spec.Environment.GetId(),
			Phase:        forwarder.Status.GetPhase(),
		})
	} else {
		table.Add(&dnsForwarderSerializedOut{
			Id:           forwarder.GetId(),
			Name:         forwarder.Spec.GetDisplayName(),
			Domains:      forwarder.Spec.GetDomains(),
			DnsServerIps: forwarder.Spec.Config.NetworkingV1ForwardViaIp.GetDnsServerIps(),
			Gateway:      forwarder.Spec.Gateway.GetId(),
			Environment:  forwarder.Spec.Environment.GetId(),
			Phase:        forwarder.Status.GetPhase(),
		})
	}

	return table.PrintWithAutoWrap(false)
}
