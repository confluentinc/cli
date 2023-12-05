package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newComputePoolListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink compute pools.",
		Args:  cobra.NoArgs,
		RunE:  c.computePoolList,
	}

	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) computePoolList(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	computePools, err := c.V2Client.ListFlinkComputePools(environmentId, region)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, computePool := range computePools {
		list.Add(&computePoolOut{
			IsCurrent:  computePool.GetId() == c.Context.GetCurrentFlinkComputePool(),
			Id:         computePool.GetId(),
			Name:       computePool.Spec.GetDisplayName(),
			CurrentCfu: computePool.Status.GetCurrentCfu(),
			MaxCfu:     computePool.Spec.GetMaxCfu(),
			Cloud:      computePool.Spec.GetCloud(),
			Region:     computePool.Spec.GetRegion(),
			Status:     computePool.Status.GetPhase(),
		})
	}
	return list.Print()
}
