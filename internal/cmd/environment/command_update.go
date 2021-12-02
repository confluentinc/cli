package environment

import (
	"context"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing Confluent Cloud environment.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.update),
	}

	cmd.Flags().String("name", "", "New name for Confluent Cloud environment.")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	id := args[0]

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	account := &orgv1.Account{
		Id:             id,
		Name:           name,
		OrganizationId: c.State.Auth.Account.OrganizationId,
	}

	if err := c.Client.Account.Update(context.Background(), account); err != nil {
		return err
	}

	utils.ErrPrintf(cmd, errors.UpdateSuccessMsg, "name", "environment", id, name)
	return nil
}
