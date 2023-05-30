package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newComputePoolListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink compute pools.",
		Args:  cobra.NoArgs,
		RunE:  c.computePoolList,
	}

	c.addRegionFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) computePoolList(cmd *cobra.Command, _ []string) error {
	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	computePools, err := c.V2Client.ListFlinkComputePools(environment, region)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, computePool := range computePools {
		list.Add(&computePoolOut{
			IsCurrent: computePool.GetId() == c.Context.GetCurrentFlinkComputePool(),
			Id:        computePool.GetId(),
			Name:      computePool.Spec.GetDisplayName(),
			Region:    computePool.Spec.GetRegion(),
		})
	}
	return list.Print()
}
