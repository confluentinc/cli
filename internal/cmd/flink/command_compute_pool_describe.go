package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *computePoolCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Flink compute pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *computePoolCommand) describe(cmd *cobra.Command, args []string) error {
	computePool, err := c.V2Client.DescribeFlinkComputePool(args[0], c.EnvironmentId())
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		Id:   computePool.GetId(),
		Name: computePool.Spec.GetDisplayName(),
	})
	return table.Print()
}
