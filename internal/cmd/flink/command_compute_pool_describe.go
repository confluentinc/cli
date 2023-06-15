package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newComputePoolDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe [id]",
		Short:             "Describe a Flink compute pool.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validComputePoolArgs),
		RunE:              c.computePoolDescribe,
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) computePoolDescribe(cmd *cobra.Command, args []string) error {
	id := c.Context.GetCurrentFlinkComputePool()
	if len(args) > 0 {
		id = args[0]
	}
	if id == "" {
		return errors.NewErrorWithSuggestions("no compute pool selected", "Select a compute pool with `confluent flink compute-pool use` or as an argument.")
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	computePool, err := c.V2Client.DescribeFlinkComputePool(id, environmentId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&computePoolOut{
		IsCurrent: computePool.GetId() == c.Context.GetCurrentFlinkComputePool(),
		Id:        computePool.GetId(),
		Name:      computePool.Spec.GetDisplayName(),
		Cfu:       computePool.Spec.GetMaxCfu(),
		Region:    computePool.Spec.GetRegion(),
		Status:    computePool.Status.GetPhase(),
	})
	return table.Print()
}
