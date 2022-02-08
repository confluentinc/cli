package iam

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/iam"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *serviceAccountCommand) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a service account.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.delete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete service account "sa-123456".`,
				Code: "confluent service-account delete sa-123456",
			},
		),
	}
}

func (c *serviceAccountCommand) delete(cmd *cobra.Command, args []string) error {
	if !strings.HasPrefix(args[0], "sa-") {
		return errors.New(errors.BadServiceAccountIDErrorMsg)
	}
	// user := &orgv1.User{ResourceId: args[0]}
	// if err := c.Client.User.DeleteServiceAccount(context.Background(), user); err != nil {
	// 	return err
	// }
	_, err := iam.DeleteIamServiceAccount(*c.IamClient, args[0], c.AuthToken())
	if err != nil {
		return err
	}
	utils.ErrPrintf(cmd, errors.DeletedServiceAccountMsg, args[0])
	return nil
}
