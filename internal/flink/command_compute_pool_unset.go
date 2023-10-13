package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newComputePoolUnsetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset",
		Short: "Unset the current Flink compute pool.",
		Long:  "Unset the current Flink compute pool that was set with `confluent flink compute-pool use`.",
		Args:  cobra.NoArgs,
		RunE:  c.unset,
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
