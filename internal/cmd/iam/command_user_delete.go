package iam

import (
	"context"
	"fmt"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
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
	resourceId := args[0]

	if ok := strings.HasPrefix(resourceId, "u-"); !ok {
		return errors.New(errors.BadResourceIDErrorMsg)
	}

	if err := c.Client.User.Delete(context.Background(), &orgv1.User{ResourceId: resourceId}); err != nil {
		return err
	}

	utils.Println(cmd, fmt.Sprintf(errors.DeletedUserMsg, resourceId))
	return nil
}
