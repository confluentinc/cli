package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *identityProviderCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete an identity provider.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete identity provider "op-12345":`,
				Code: "confluent iam provider delete op-12345",
			},
		),
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *identityProviderCommand) delete(cmd *cobra.Command, args []string) error {
	provider, err := c.V2Client.GetIdentityProvider(args[0])
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.IdentityProvider, args[0], provider.GetDisplayName())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, provider.GetDisplayName()); err != nil {
		return err
	}

	err = c.V2Client.DeleteIdentityProvider(args[0])
	if err != nil {
		return err
	}

	utils.ErrPrintf(cmd, errors.DeletedResourceMsg, resource.IdentityProvider, args[0])
	return nil
}
