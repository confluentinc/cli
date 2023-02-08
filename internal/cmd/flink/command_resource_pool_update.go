package flink

import (
	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *resourcePoolCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a Flink resource pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *resourcePoolCommand) update(cmd *cobra.Command, args []string) error {
	resourcePoolUpdate := flinkv2.FrpmV2ResourcePoolUpdate{
		Spec: &flinkv2.FrpmV2ResourcePoolSpecUpdate{
			// TODO: CSUs
		},
	}

	resourcePool, err := c.V2Client.UpdateFlinkResourcePool(args[0], resourcePoolUpdate)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		Id: resourcePool.GetId(),
	})
	return table.Print()
}
