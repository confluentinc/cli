package flink

import (
	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newComputePoolUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "update <id>",
		Short:  "Update a Flink compute pool.",
		Args:   cobra.ExactArgs(1),
		RunE:   c.update,
		Hidden: true, // TODO: Remove for GA
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	computePoolUpdate := flinkv2.FcpmV2ComputePoolUpdate{
		Spec: &flinkv2.FcpmV2ComputePoolSpecUpdate{},
	}

	computePool, err := c.V2Client.UpdateFlinkComputePool(args[0], computePoolUpdate)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&computePoolOut{
		Id: computePool.GetId(),
	})
	return table.Print()
}
