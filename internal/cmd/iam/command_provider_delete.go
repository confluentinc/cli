package iam

import (
	"fmt"

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
	pcmd.AddSkipInvalidFlag(cmd)

	return cmd
}

func (c *identityProviderCommand) delete(cmd *cobra.Command, args []string) error {
	displayName, validArgs, err := c.validateArgs(cmd, args)
	if err != nil {
		return err
	}
	args = validArgs

	if _, err := form.ConfirmDeletionType(cmd, resource.IdentityProvider, displayName, args); err != nil {
		return err
	}

	var errs error
	var deleted []string
	for _, id := range args {
		if err := c.V2Client.DeleteIdentityProvider(id); err != nil {
			errs = errors.Join(errs, err)
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.IdentityProvider)

	return errs
}

func (c *identityProviderCommand) validateArgs(cmd *cobra.Command, args []string) (string, []string, error) {
	var displayName string
	describeFunc := func(id string) error {
		if provider, err := c.V2Client.GetIdentityProvider(id); err != nil {
			return err
		} else if id == args[0] {
			displayName = provider.GetDisplayName()
		}
		return nil
	}

	validArgs, err := deletion.ValidateArgsForDeletion(cmd, args, resource.IdentityProvider, describeFunc)
	err = errors.NewWrapAdditionalSuggestions(err, fmt.Sprintf(errors.ListResourceSuggestions, resource.IdentityProvider, "iam provider"))

	return displayName, validArgs, err
}
