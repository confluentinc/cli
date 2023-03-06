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

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	environment, httpResp, err := c.V2Client.GetOrgEnvironment(id)
	if err != nil {
		return errors.CatchOrgV2ResourceNotFoundError(err, resource.Environment, httpResp)
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.Environment, id, *environment.DisplayName)
	if _, err := form.ConfirmDeletion(cmd, promptMsg, *environment.DisplayName); err != nil {
		return err
	}

	httpResp, err = c.V2Client.DeleteOrgEnvironment(id)
	if err != nil {
		return errors.CatchOrgV2ResourceNotFoundError(err, resource.Environment, httpResp)
	}

	output.ErrPrintf(errors.DeletedResourceMsg, resource.Environment, id)
	if id == c.EnvironmentId() {
		c.Context.SetEnvironment(nil)

		if err := c.Config.Save(); err != nil {
			return errors.Wrap(err, errors.EnvSwitchErrorMsg)
		}
	}
	return nil
}
