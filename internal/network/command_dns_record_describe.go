package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newDnsRecordDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a DNS record.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validDnsRecordArgs),
		RunE:              c.dnsRecordDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe DNS recorder "dnsrec-123456".`,
				Code: "confluent network dns recorder describe dnsrec-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) dnsRecordDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	record, err := c.V2Client.GetDnsRecord(environmentId, args[0])
	if err != nil {
		return err
	}

	return printDnsRecordTable(cmd, record)
}
