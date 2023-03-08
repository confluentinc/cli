package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *computePoolCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "delete <id>",
		Short:  "Delete a Flink compute pool.",
		Args:   cobra.ExactArgs(1),
		RunE:   c.delete,
		Hidden: true, // TODO: Remove for GA
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *computePoolCommand) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	computePool, err := c.V2Client.DescribeFlinkComputePool(id, c.EnvironmentId())
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.FlinkComputePool, computePool.Metadata.GetResourceName(), computePool.GetId())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, computePool.GetId()); err != nil {
		return err
	}

	if err := c.V2Client.DeleteFlinkComputePool(id); err != nil {
		return err
	}

	output.ErrPrintf(errors.DeletedResourceMsg, resource.FlinkComputePool, id)
	return nil
}
