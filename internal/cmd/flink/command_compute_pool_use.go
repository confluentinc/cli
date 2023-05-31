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
		Use:   "use <id>",
		Short: "Choose a Flink compute pool to be used in subsequent commands.",
		Long:  "Choose a Flink compute pool to be used in subsequent commands which support passing a compute pool with the `--compute-pool` flag.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.computePoolUse,
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) computePoolUse(cmd *cobra.Command, args []string) error {
	id := args[0]
	if _, err := c.V2Client.DescribeFlinkComputePool(id, c.Context.GetCurrentEnvironment()); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), "List available compute pools with `confluent flink compute-pool list`.")
	}

	if err := c.Context.SetCurrentFlinkComputePool(id); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(errors.UsingResourceMsg, resource.FlinkComputePool, args[0])
	return nil
}
