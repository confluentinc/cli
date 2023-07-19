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

func (c *command) newComputePoolDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a Flink compute pool.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validComputePoolArgs),
		RunE:              c.computePoolDelete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) computePoolDelete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	computePool, err := c.V2Client.DescribeFlinkComputePool(args[0], environmentId)
	if err != nil {
		if coder, ok := err.(errors.Coder); ok {
			coder.GetStatusCode()
		}
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.FlinkComputePool, computePool.GetId(), computePool.Spec.GetDisplayName())
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, computePool.Spec.GetDisplayName()); err != nil || !ok {
		return err
	}

	if err := c.V2Client.DeleteFlinkComputePool(args[0], environmentId); err != nil {
		return err
	}

	output.Printf(errors.DeletedResourceMsg, resource.FlinkComputePool, args[0])

	if computePool.GetId() == c.Context.GetCurrentFlinkComputePool() {
		if err := c.Context.SetCurrentFlinkComputePool(""); err != nil {
			return err
		}
		if err := c.Config.Save(); err != nil {
			return err
		}
	}

	return nil
}
