package customcodelogging

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
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
				Code: "confluent custom-code-logging describe ccl-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	return cmd
}

func (c *customCodeLoggingCommand) describe(cmd *cobra.Command, args []string) error {
	customCodeLogging, err := c.V2Client.DescribeCustomCodeLogging(args[0])
	if err != nil {
		return err
	}
	return printTable(cmd, customCodeLogging)
}
