package iam

import (
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *identityProviderCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an identity provider.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an identity provider named "demo-identity-provider".`,
				Code: `confluent iam provider create demo-identity-provider --description "new description" --jwks-uri https://company.provider.com/oauth2/v1/keys --issuer-uri https://company.provider.com`,
			},
		),
	}

	cmd.Flags().String("issuer-uri", "", "URI of the identity provider issuer.")
	cmd.Flags().String("jwks-uri", "", "JWKS (JSON Web Key Set) URI of the identity provider.")
	cmd.Flags().String("description", "", "Description of the identity provider.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("issuer-uri"))
	cobra.CheckErr(cmd.MarkFlagRequired("jwks-uri"))

	return cmd
}

func (c *identityProviderCommand) create(cmd *cobra.Command, args []string) error {
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	issuerUri, err := cmd.Flags().GetString("issuer-uri")
	if err != nil {
		return err
	}

	jwksUri, err := cmd.Flags().GetString("jwks-uri")
	if err != nil {
		return err
	}

	createIdentityProvider := identityproviderv2.IamV2IdentityProvider{
		DisplayName: identityproviderv2.PtrString(args[0]),
		Description: identityproviderv2.PtrString(description),
		Issuer:      identityproviderv2.PtrString(issuerUri),
		JwksUri:     identityproviderv2.PtrString(jwksUri),
	}
	provider, err := c.V2Client.CreateIdentityProvider(createIdentityProvider)
	if err != nil {
		return err
	}

	return printIdentityProvider(cmd, provider)
}
