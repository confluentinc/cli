package iam

import (
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *identityPoolCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete an identity pool.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete identity pool pool-12345.`,
				Code: "confluent iam pool delete pool-12345",
			},
		),
	}

	cmd.Flags().String("provider", "", "ID of this pool's identity provider.")
	_ = cmd.MarkFlagRequired("provider")

	return cmd
}

func (c *identityPoolCommand) delete(cmd *cobra.Command, args []string) error {

	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	_, err = c.V2Client.DeleteIdentityPool(args[0], provider)
	if err != nil {
		return errors.Errorf(`failed to delete identity pool "%s": %v`, args[0], err)
	}

	utils.ErrPrintf(cmd, errors.DeletedIdentityPoolMsg, args[0])
	return nil
}
