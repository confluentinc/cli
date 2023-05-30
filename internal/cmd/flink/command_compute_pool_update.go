package flink

import (
	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink/v2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newComputePoolUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a Flink compute pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.computePoolUpdate,
	}

	cmd.Flags().Int32("cfu", 1, "Number of Confluent Flink Units (CFU).")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) computePoolUpdate(cmd *cobra.Command, args []string) error {
	id := c.Context.GetCurrentFlinkComputePool()
	if len(args) > 0 {
		id = args[0]
	}

	cfu, err := cmd.Flags().GetInt32("cfu")
	if err != nil {
		return err
	}

	update := flinkv2.FcpmV2ComputePoolUpdate{
		Id:   flinkv2.PtrString(id),
		Spec: &flinkv2.FcpmV2ComputePoolSpecUpdate{MaxCfu: flinkv2.PtrInt32(cfu)},
	}

	computePool, err := c.V2Client.UpdateFlinkComputePool(id, update)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&computePoolOut{
		IsCurrent: computePool.GetId() == c.Context.GetCurrentFlinkComputePool(),
		Id:        computePool.GetId(),
		Name:      computePool.Spec.GetDisplayName(),
		Region:    computePool.Spec.GetRegion(),
	})
	return table.Print()
}
