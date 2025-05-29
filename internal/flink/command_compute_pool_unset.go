package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newComputePoolUnsetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "unset",
		Short:       "Unset the current Flink compute pool.",
		Long:        "Unset the current Flink compute pool that was set with the `use` command.",
		Args:        cobra.NoArgs,
		RunE:        c.unset,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Unset default compute pool:`,
				Code: "confluent flink compute-pool unset",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) unset(_ *cobra.Command, args []string) error {
	computePoolToUnset := c.Context.GetCurrentFlinkComputePool()
	if computePoolToUnset == "" {
		return nil
	}

	if err := c.Context.SetCurrentFlinkComputePool(""); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UnsetResourceMsg, resource.FlinkComputePool, computePoolToUnset)
	return nil
}
