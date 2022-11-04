package environment

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a Confluent Cloud environment and all its resources.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}
	cmd.Flags().Bool("force", false, "Skip the deletion confirmation prompt.")

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	environment, httpResp, err := c.V2Client.GetOrgEnvironment(id)
	if err != nil {
		return errors.CatchEnvironmentNotFoundError(err, httpResp)
	}
	err = form.ConfirmDeletion(cmd, resource.Environment, id, *environment.DisplayName)
	if err != nil {
		return err
	}

	httpResp, err = c.V2Client.DeleteOrgEnvironment(id)
	if err != nil {
		return errors.CatchEnvironmentNotFoundError(err, httpResp)
	}

	utils.ErrPrintf(cmd, errors.DeletedResourceMsg, resource.Environment, id)
	return nil
}
