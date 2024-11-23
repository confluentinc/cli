package customcodelogging

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *customCodeLoggingCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a custom code logging.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe custom code logging.",
				Code: "confluent custom-code-logging describe ccl-123456 --environment env-000000",
			},
		),
	}
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	return cmd
}

func (c *customCodeLoggingCommand) describe(cmd *cobra.Command, args []string) error {
	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	customCodeLogging, err := c.V2Client.DescribeCustomCodeLogging(args[0], environment)
	if err != nil {
		return err
	}
	table := output.NewTable(cmd)
	table.Add(getCustomCodeLogging(customCodeLogging))
	table.Filter([]string{"ID", "Cloud", "Region", "Environment", "Topic", "Cluster", "LogLevel"})
	return table.Print()
}
