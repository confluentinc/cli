package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *resourcePoolCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink resource pools.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *resourcePoolCommand) list(cmd *cobra.Command, args []string) error {
	resourcePools, err := c.V2Client.ListFlinkResourcePools()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, resourcePool := range resourcePools {
		list.Add(&out{
			Id: resourcePool.GetId(),
		})
	}
	return list.Print()
}
