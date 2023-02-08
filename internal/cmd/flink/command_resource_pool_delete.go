package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *resourcePoolCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a Flink resource pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *resourcePoolCommand) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	resourcePool, err := c.V2Client.DescribeFlinkResourcePool(id)
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.FlinkResourcePool, resourcePool.Metadata.GetResourceName(), resourcePool.GetId())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, resourcePool.GetId()); err != nil {
		return err
	}

	if err := c.V2Client.DeleteFlinkResourcePool(id); err != nil {
		return err
	}

	utils.ErrPrintf(cmd, errors.DeletedResourceMsg, resource.Environment, id)
	return nil
}
