package configuration

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newDescribeCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "describe [key]",
		Short:             "Describe a user-configurable field.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `See if update checks are enabled by describing "disable_update_check".`,
				Code: "confluent configuration describe disable_update_check",
			},
		),
	}
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	field := args[0]
	if _, ok := c.configWhiteList[field]; !ok {
		return fmt.Errorf(fieldNotConfigurableError, field)
	}
	table := output.NewTable(cmd)
	table.Add(c.newConfigurationOut(field))
	return table.Print()
}
