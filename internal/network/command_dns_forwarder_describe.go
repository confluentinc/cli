package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newDnsForwarderDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a DNS forwarder.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validDnsForwarderArgs),
		RunE:              c.dnsForwarderDescribe,

		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe DNS forwarder "dnsf-123456".`,
				Code: "confluent network dns-forwarder describe dnsf-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) dnsForwarderDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	forwarder, err := c.V2Client.GetDnsForwarder(environmentId, args[0])
	if err != nil {
		return err
	}

	return printDnsForwarderTable(cmd, forwarder)
}
