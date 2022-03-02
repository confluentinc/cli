package environment

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/org"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a Confluent Cloud environment and all its resources.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.delete),
	}
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	_, err := org.DeleteOrgEnvironment(c.V2Client.OrgClient, id, c.AuthToken())
	if err != nil {
		return err
	}

	utils.ErrPrintf(cmd, errors.DeletedEnvMsg, id)
	return nil
}
