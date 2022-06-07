package iam

import (
	"fmt"
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/identity-provider/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *identityProviderCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a service account.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the description of identity provider "op-123456".`,
				Code: `confluent iam provider update op-123456 --description "Update demo identity provider information."`,
			},
		),
	}

	cmd.Flags().String("description", "", "Description of the identity provider.")
	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func (c *identityProviderCommand) update(cmd *cobra.Command, args []string) error {
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if err := requireLen(description, descriptionLength, "description"); err != nil {
		return err
	}

	if resource.LookupType(args[0]) != resource.IdentityProvider {
		return fmt.Errorf(errors.BadResourceIDErrorMsg, resource.IdentityProviderPrefix)
	}
	identityProviderId := args[0]

	update := identityproviderv2.IamV2IdentityProviderUpdate{
		Description: &description,
	}
	_, httpresp, err := c.V2Client.UpdateIdentityProvider(args[0], update)
	if err != nil {
		return errors.CatchIdentityProviderNotFoundError(err, httpresp, identityProviderId)
	}

	utils.ErrPrintf(cmd, errors.UpdateSuccessMsg, "description", "identity provider", identityProviderId, description)
	return nil
}
