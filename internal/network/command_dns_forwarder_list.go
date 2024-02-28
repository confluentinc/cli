package network

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newDnsForwarderListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS forwarders.",
		Args:  cobra.NoArgs,
		RunE:  c.dnsForwarderList,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) dnsForwarderList(cmd *cobra.Command, _ []string) error {
	forwarders, err := c.getDnsForwarders()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, forwarder := range forwarders {
		if forwarder.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if forwarder.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}
		sort.Strings(forwarder.Spec.GetDomains())

		if output.GetFormat(cmd) == output.Human {
			list.Add(&dnsForwarderHumanOut{
				Id:           forwarder.GetId(),
				Name:         forwarder.Spec.GetDisplayName(),
				Domains:      strings.Join(forwarder.Spec.GetDomains(), ", "),
				DnsServerIps: strings.Join(forwarder.Spec.Config.NetworkingV1ForwardViaIp.GetDnsServerIps(), ", "),
				Gateway:      forwarder.Spec.Gateway.GetId(),
				Environment:  forwarder.Spec.Environment.GetId(),
				Phase:        forwarder.Status.GetPhase(),
			})
		} else {
			list.Add(&dnsForwarderSerializedOut{
				Id:           forwarder.GetId(),
				Name:         forwarder.Spec.GetDisplayName(),
				Domains:      forwarder.Spec.GetDomains(),
				DnsServerIps: forwarder.Spec.Config.NetworkingV1ForwardViaIp.GetDnsServerIps(),
				Gateway:      forwarder.Spec.Gateway.GetId(),
				Environment:  forwarder.Spec.Environment.GetId(),
				Phase:        forwarder.Status.GetPhase(),
			})
		}
	}

	return list.Print()
}
