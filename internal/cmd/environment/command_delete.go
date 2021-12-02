package environment

import (
	"context"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
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
		RunE:              pcmd.NewCLIRunE(c.delete),
	}
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	account := &orgv1.Account{Id: id, OrganizationId: c.State.Auth.Account.OrganizationId}

	if err := c.Client.Account.Delete(context.Background(), account); err != nil {
		return err
	}

	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, id)

	utils.ErrPrintf(cmd, errors.DeletedEnvMsg, id)
	return nil
}
