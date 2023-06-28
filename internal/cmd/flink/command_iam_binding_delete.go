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

func (c *command) newIamBindingDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a Flink IAM binding.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.iamBindingDelete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) iamBindingDelete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.FlinkIamBinding, args[0], args[0])
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, args[0]); err != nil || !ok {
		return err
	}

	if err := c.V2Client.DeleteFlinkIAMBinding(args[0], environmentId); err != nil {
		return err
	}

	output.Printf(errors.DeletedResourceMsg, resource.FlinkIamBinding, args[0])

	return nil
}
