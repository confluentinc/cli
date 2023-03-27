package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
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
				Text: `Delete identity pool "pool-12345":`,
				Code: "confluent iam pool delete pool-12345 --provider op-12345",
			},
		),
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("provider"))

	return cmd
}

func (c *identityPoolCommand) delete(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	pool, err := c.V2Client.GetIdentityPool(args[0], provider)
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.IdentityPool, args[0], pool.GetDisplayName())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, pool.GetDisplayName()); err != nil {
		return err
	}

	if err := c.V2Client.DeleteIdentityPool(args[0], provider); err != nil {
		return err
	}

	output.ErrPrintf(errors.DeletedResourceMsg, resource.IdentityPool, args[0])
	return nil
}
