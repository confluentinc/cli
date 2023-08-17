package configuration

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <key>",
		Short:             "Describe a user-configurable field.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `View the "disable_update_check" configuration.`,
				Code: "confluent configuration describe disable_update_check",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	whitelist := getWhitelist(c.cfg)
	field := args[0]

	if _, ok := whitelist[field]; !ok {
		return fmt.Errorf(fieldNotConfigurableError, field)
	}

	table := output.NewTable(cmd)
	table.Add(c.newFieldOut(field, whitelist))
	return table.Print()
}
