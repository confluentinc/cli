package iam

import (
	"errors"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	perrors "github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *identityProviderCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-N]",
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

	return cmd
}

func (c *identityProviderCommand) delete(cmd *cobra.Command, args []string) error {
	displayName, err := c.checkExistence(cmd, args)
	if err != nil {
		return err
	}

	if _, err := form.ConfirmDeletionType(cmd, resource.IdentityProvider, displayName, args); err != nil {
		return err
	}

	var errs error
	for _, providerId := range args {
		if err := c.V2Client.DeleteIdentityProvider(providerId); err != nil {
			errs = errors.Join(errs, err)
		} else {
			output.ErrPrintf(perrors.DeletedResourceMsg, resource.IdentityProvider, providerId)
		}
	}

	return errs
}

func (c *identityProviderCommand) checkExistence(cmd *cobra.Command, args []string) (string, error) {
	// Single
	if len(args) == 1 {
		if provider, err := c.V2Client.GetIdentityProvider(args[0]); err != nil {
			return "", err
		} else {
			return provider.GetDisplayName(), nil
		}
	}

	// Multiple
	identityProviders, err := c.V2Client.ListIdentityProviders()
	if err != nil {
		return "", err
	}

	providerSet := types.NewSet()
	for _, provider := range identityProviders {
		providerSet.Add(provider.GetId())
	}

	invalidProviders := providerSet.Difference(args)
	if len(invalidProviders) > 0 {
		return "", perrors.New("provider(s) not found or access forbidden: " + utils.ArrayToCommaDelimitedStringWithAnd(invalidProviders))
	}

	return "", nil
}
