package environment

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a Confluent Cloud environment and all its resources.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	httpResp, err := c.V2Client.DeleteOrgEnvironment(id)
	if err != nil {
		return errors.CatchEnvironmentNotFoundError(err, httpResp)
	}

	utils.ErrPrintf(cmd, errors.DeletedEnvMsg, id)
	return nil
}
