package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newComputePoolDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe [id]",
		Short: "Describe a Flink compute pool.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.computePoolDescribe,
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) computePoolDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	computePool, err := c.V2Client.DescribeFlinkComputePool(c.Context.GetCurrentFlinkComputePool(), environmentId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&computePoolOut{
		Current: computePool.GetId() == c.Context.GetCurrentFlinkComputePool(),
		Id:      computePool.GetId(),
		Name:    computePool.Spec.GetDisplayName(),
	})
	return table.Print()
}
