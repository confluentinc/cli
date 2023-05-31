package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newComputePoolUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id>",
		Short:             "Choose a Flink compute pool to be used in subsequent commands.",
		Long:              "Choose a Flink compute pool to be used in subsequent commands which support passing a compute pool with the `--compute-pool` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.computePoolUse,
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) computePoolUse(cmd *cobra.Command, args []string) error {
	if err := c.Context.SetCurrentFlinkComputePool(args[0]); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(errors.UsingResourceMsg, resource.FlinkComputePool, args[0])
	return nil
}
