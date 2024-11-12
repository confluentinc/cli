package customcodelogging

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *customCodeLoggingCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom code loggings.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List custom code loggings in the org.",
				Code: "confluent custom-code-logging list --environment env-000000",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	return cmd
}

func (c *customCodeLoggingCommand) list(cmd *cobra.Command, _ []string) error {
	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	customCodeLoggings, err := c.V2Client.ListCustomCodeLoggings(environment, environment)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, customCodeLogging := range customCodeLoggings {
		list.Add(getCustomCodeLogging(customCodeLogging))
	}
	list.Filter([]string{"Id", "Cloud", "Region", "Environment"})
	return list.Print()
}
