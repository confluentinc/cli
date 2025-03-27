package iam

import (
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
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
				Code: `confluent iam provider update op-123456 --description "updated description"`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the identity provider.")
	cmd.Flags().String("description", "", "Description of the identity provider.")
	cmd.Flags().String("identity-claim", "", "The JSON Web Token (JWT) claim to extract the authenticating identity to Confluent resources from "+
		"[Registered Claim Names](https://datatracker.ietf.org/doc/html/rfc7519#section-4.1). This appears "+
		"in audit log records. Note: if the client specifies mapping to one identity pool ID, the identity "+
		"claim configured with that pool will be used instead.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("name", "description", "identity-claim")

	return cmd
}

func (c *identityProviderCommand) update(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	identityClaim, err := cmd.Flags().GetString("identity-claim")
	if err != nil {
		return err
	}

	update := identityproviderv2.IamV2IdentityProvider{Id: identityproviderv2.PtrString(args[0])}
	if name != "" {
		update.DisplayName = identityproviderv2.PtrString(name)
	}
	if identityClaim != "" {
		update.IdentityClaim = identityproviderv2.PtrString(identityClaim)
	}
	if description != "" {
		update.Description = identityproviderv2.PtrString(description)
	}

	provider, err := c.V2Client.UpdateIdentityProvider(update)
	if err != nil {
		return err
	}

	return printIdentityProvider(cmd, provider)
}
