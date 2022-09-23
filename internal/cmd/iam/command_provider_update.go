package iam

import (
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
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

	cmd.Flags().String("description", "", "Description of the identity provider.")
	cmd.Flags().String("name", "", "Name of the identity provider.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *identityProviderCommand) update(cmd *cobra.Command, args []string) error {
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	if description == "" && name == "" {
		return errors.New(errors.IdentityProviderNoOpUpdateErrorMsg)
	}

	identityProviderId := args[0]
	update := identityproviderv2.IamV2IdentityProviderUpdate{Id: &identityProviderId}
	if name != "" {
		update.DisplayName = &name
	}
	if description != "" {
		update.Description = &description
	}

	identityProvider, err := c.V2Client.UpdateIdentityProvider(update)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&identityProviderOut{
		Id:          identityProvider.GetId(),
		Name:        identityProvider.GetDisplayName(),
		Description: identityProvider.GetDescription(),
		IssuerUri:   identityProvider.GetIssuer(),
		JwksUri:     identityProvider.GetJwksUri(),
	})
	return table.Print()
}
