package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newComputePoolUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "use <id>",
		Short:             "Use a Flink compute pool in subsequent commands.",
		Long:              "Choose a Flink compute pool to be used in subsequent commands which support passing a compute pool with the `--compute-pool` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validComputePoolArgs),
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		RunE:              c.computePoolUse,
	}
}

func (c *command) computePoolUse(_ *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	if _, err := c.V2Client.DescribeFlinkComputePool(args[0], environmentId); err != nil {
		return err
	}

	if err := c.Context.SetCurrentFlinkComputePool(args[0]); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UsingResourceMsg, resource.FlinkComputePool, args[0])
	return nil
}
