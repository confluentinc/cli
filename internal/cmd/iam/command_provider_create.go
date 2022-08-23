package iam

import (
	"github.com/spf13/cobra"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *identityProviderCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an identity provider.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an identity provider named "DemoIdentityProvider".`,
				Code: `confluent iam provider create DemoIdentityProvider --description "description of provider" --jwks-uri https://company.provider.com/oauth2/v1/keys --issuer-uri https://company.provider.com`,
			},
		),
	}

	cmd.Flags().String("issuer-uri", "", "URI of the identity provider issuer.")
	cmd.Flags().String("jwks-uri", "", "JWKS (JSON Web Key Set) URI of the identity provider.")
	cmd.Flags().String("description", "", "Description of the identity provider.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("issuer-uri")
	_ = cmd.MarkFlagRequired("jwks-uri")

	return cmd
}

func (c *identityProviderCommand) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	issuer, err := cmd.Flags().GetString("issuer-uri")
	if err != nil {
		return err
	}

	jwksuri, err := cmd.Flags().GetString("jwks-uri")
	if err != nil {
		return err
	}

	newIdentityProvider := identityproviderv2.IamV2IdentityProvider{
		DisplayName: identityproviderv2.PtrString(name),
		Description: identityproviderv2.PtrString(description),
		Issuer:      identityproviderv2.PtrString(issuer),
		JwksUri:     identityproviderv2.PtrString(jwksuri),
	}
	resp, httpResp, err := c.V2Client.CreateIdentityProvider(newIdentityProvider)
	if err != nil {
		return errors.CatchV2ErrorMessageWithResponse(err, httpResp)
	}

	identityProvider := &identityProvider{
		Id:        *resp.Id,
		Name:      *resp.DisplayName,
		IssuerUri: *resp.Issuer,
		JwksUri:   *resp.JwksUri,
	}
	if resp.Description != nil {
		identityProvider.Description = *resp.Description
	}

	return output.DescribeObject(cmd, identityProvider, providerListFields, providerHumanLabelMap, providerStructuredLabelMap)
}
