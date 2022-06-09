package iam

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/identity-provider/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *identityProviderCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an identity provider.",
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

	cmd.Flags().String("name", "", "Name of the identity provider.")
	cmd.Flags().String("description", "", "Description of the identity provider.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *identityProviderCommand) update(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	if err := requireLen(name, nameLength, "name"); err != nil {
		return err
	}

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
		Id: &identityProviderId,
	}
	if name != "" {
		update.DisplayName = &name
	}
	if description != "" {
		update.Description = &description
	}

	resp, httpresp, err := c.V2Client.UpdateIdentityProvider(update)
	if err != nil {
		return errors.CatchIdentityProviderNotFoundError(err, httpresp, identityProviderId)
	}

	describeIdentityProvider := &identityProvider{Id: *resp.Id, DisplayName: *resp.DisplayName, Description: *resp.Description, Issuer: *resp.Issuer, JwksUri: *resp.JwksUri}
	return output.DescribeObject(cmd, describeIdentityProvider, providerListFields, providerHumanLabelMap, providerStructuredLabelMap)
}
