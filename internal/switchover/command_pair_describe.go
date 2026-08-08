package switchover

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newPairDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a switchover pair.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPairArgs),
		RunE:              c.pairDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe switchover pair "sw-123456".`,
				Code: "confluent switchover pair describe sw-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) pairDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	pair, err := c.V2Client.DescribeSwitchoverPair(args[0], environmentId)
	if err != nil {
		return err
	}

	return printPairTable(cmd, pair)
}
