package iam

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *identityProviderCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete identity providers.",
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
	if err := c.confirmDeletion(cmd, args); err != nil {
		return err
	}

	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var deleted []string
	for _, id := range args {
		if err := c.V2Client.DeleteIdentityProvider(id); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.IdentityProvider)

	return errs.ErrorOrNil()
}

func (c *identityProviderCommand) confirmDeletion(cmd *cobra.Command, args []string) error {
	var displayName string
	describeFunc := func(id string) error {
		provider, err := c.V2Client.GetIdentityProvider(id)
		if err == nil && id == args[0] {
			displayName = provider.GetDisplayName()
		}
		return err
	}

	if err := deletion.ValidateArgsForDeletion(cmd, args, resource.IdentityProvider, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.IdentityProvider, args[0], displayName); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.IdentityProvider, args); err != nil || !ok {
			return err
		}
	}

	return nil
}
