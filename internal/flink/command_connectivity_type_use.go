package flink

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:        "use <region-access>",
		Short:      "Select a Flink connectivity type.",
		Long:       "Select a Flink connectivity type for the current environment as \"public\" or \"private\". If unspecified, the CLI will default to the connectivity type that was set at the organization level.",
		Args:       cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs:  fields,
		RunE:       c.ConnectivityTypeUse,
		Deprecated: `please run "confluent flink endpoint list" and "confluent flink endpoint use" to specify the Flink client endpoint`,
	}
}

func (c *command) ConnectivityTypeUse(_ *cobra.Command, args []string) error {
	if err := c.Context.SetCurrentFlinkAccessType(args[0]); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}
	output.Printf(c.Config.EnableColor, errors.UsingResourceMsg, resource.FlinkConnectivityType, args[0])
	return nil
}
