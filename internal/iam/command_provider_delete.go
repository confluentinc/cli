package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/resource"
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

	pcmd.AddContextFlag(cmd, c.CLICommand)
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
	existenceFunc := func(id string) bool {
		provider, err := c.V2Client.GetIdentityProvider(id)
		if err != nil {
			return false
		}
		if id == args[0] {
			displayName = provider.GetDisplayName()
		}

		return true
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.IdentityProvider, existenceFunc); err != nil {
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
