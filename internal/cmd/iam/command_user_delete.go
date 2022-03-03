package iam

import (
	"context"
	"fmt"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c userCommand) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a user from your organization.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
	}
}

func (c userCommand) delete(cmd *cobra.Command, args []string) error {
	if resource.LookupType(args[0]) != resource.User {
		return errors.New(errors.BadResourceIDErrorMsg)
	}
	userId := args[0]

	if err := c.Client.User.Delete(context.Background(), &orgv1.User{ResourceId: userId}); err != nil {
		return err
	}

	utils.Println(cmd, fmt.Sprintf(errors.DeletedUserMsg, userId))
	return nil
}
