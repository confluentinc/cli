package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *identityProviderCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more identity providers.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
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
	if confirm, err := c.confirmDeletion(cmd, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteIdentityProvider(id)
	}

	_, err := resource.Delete(args, deleteFunc, resource.IdentityProvider)
	return err
}

func (c *identityProviderCommand) confirmDeletion(cmd *cobra.Command, args []string) (bool, error) {
	var displayName string
	describeFunc := func(id string) error {
		provider, err := c.V2Client.GetIdentityProvider(id)
		if err != nil {
			return err
		}
		if id == args[0] {
			displayName = provider.GetDisplayName()
		}

		return nil
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.IdentityProvider, describeFunc); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.IdentityProvider, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.IdentityProvider, args[0], displayName), displayName); err != nil {
		return false, err
	}

	return true, nil
}
