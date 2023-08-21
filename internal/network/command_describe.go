package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a network.",
		Args:  cobra.ExactArgs(1),
		// TODO: Implement autocompletion after List Network is implemented.
		// ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE: c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe Confluent network "n-abcde1".`,
				Code: `confluent network describe n-abcde1`,
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	network, err := c.V2Client.GetNetwork(environmentId, args[0])
	if err != nil {
		return err
	}

	return printTable(cmd, network)
}
