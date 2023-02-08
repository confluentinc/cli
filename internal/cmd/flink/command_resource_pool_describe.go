package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *resourcePoolCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Flink resource pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *resourcePoolCommand) describe(cmd *cobra.Command, args []string) error {
	resourcePool, err := c.V2Client.DescribeFlinkResourcePool(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		Id: resourcePool.GetId(),
	})
	return table.Print()
}
