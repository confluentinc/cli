package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *computePoolCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink compute pools.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddRegionFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("region")

	return cmd
}

func (c *computePoolCommand) list(cmd *cobra.Command, args []string) error {
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	computePools, err := c.V2Client.ListFlinkComputePools(region, c.EnvironmentId())
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, computePool := range computePools {
		list.Add(&out{
			Id:   computePool.GetId(),
			Name: computePool.Spec.GetDisplayName(),
		})
	}
	return list.Print()
}
