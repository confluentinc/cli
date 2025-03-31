package flink

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "use <region-access>",
		Short:     "Select a Flink connectivity type.",
		Long:      "Select a Flink connectivity type for the current environment as \"public\" or \"private\". If unspecified, the CLI will default to public connectivity type.",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: fields,
		RunE:      c.ConnectivityTypeUse,
	}
}

func (c *command) ConnectivityTypeUse(_ *cobra.Command, args []string) error {
	warning := errors.NewWarningWithSuggestions(`The current command "confluent flink endpoint list" still works to specify a public or private Flink dataplane endpoint`, `\nAlternatively, you can run "confluent flink endpoint list" and "confluent flink endpoint use" to view and specify a custom endpoint for Flink dataplane client, including CCN endpoints.`)
	output.ErrPrint(true, warning.DisplayWarningWithSuggestions())

	if err := c.Context.SetCurrentFlinkAccessType(args[0]); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}
	output.Printf(c.Config.EnableColor, errors.UsingResourceMsg, resource.FlinkConnectivityType, args[0])
	return nil
}
