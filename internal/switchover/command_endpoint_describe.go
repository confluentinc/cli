package switchover

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newEndpointDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a switchover endpoint.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validEndpointArgs),
		RunE:              c.endpointDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe switchover endpoint "se-123456".`,
				Code: "confluent switchover endpoint describe se-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) endpointDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	endpoint, err := c.V2Client.DescribeSwitchoverEndpoint(args[0], environmentId)
	if err != nil {
		return err
	}

	return printEndpointTable(cmd, endpoint)
}
