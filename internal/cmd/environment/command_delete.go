package environment

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a Confluent Cloud environment and all its resources.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id, err := convertNameToId(args[0], c.AuthenticatedCLICommand.V2Client)
	if err != nil {
		return err
	}

	environment, err := c.V2Client.GetOrgEnvironment(id)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), "List available environments with `confluent environment list`.")
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.Environment, id, environment.GetDisplayName())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, environment.GetDisplayName()); err != nil {
		return err
	}

	if err := c.V2Client.DeleteOrgEnvironment(id); err != nil {
		return err
	}

	output.ErrPrintf(errors.DeletedResourceMsg, resource.Environment, id)

	if id == c.Context.GetCurrentEnvironment() {
		c.Context.SetCurrentEnvironment("")
		if err := c.Config.Save(); err != nil {
			return err
		}
	}

	c.Context.DeleteEnvironment(id)
	_ = c.Config.Save()

	return nil
}
