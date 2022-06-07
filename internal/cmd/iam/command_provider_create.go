package iam

import (
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/identity-provider/v2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *identityProviderCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an identity provider.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a identity provider named `DemoIdentityProvider`.",
				Code: `confluent iam provider create DemoIdentityProvider --description "description about idp" --jwks-uri https://company.provider.com/oauth2/v1/keys --issuer-uri https://company.provider.com`,
			},
		),
	}

	cmd.Flags().String("description", "", "Description of the identity provider.")
	_ = cmd.MarkFlagRequired("description")
	cmd.Flags().String("jwks-uri", "", "JWKS URI of the identity provider.")
	_ = cmd.MarkFlagRequired("jwks-uri")
	cmd.Flags().String("issuer-uri", "", "Issuer URI of the identity provider.")
	_ = cmd.MarkFlagRequired("issuer-uri")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *identityProviderCommand) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := requireLen(name, nameLength, "provider name"); err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if err := requireLen(description, descriptionLength, "description"); err != nil {
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

	createIdentityProvider := identityproviderv2.IamV2IdentityProvider{
		DisplayName: identityproviderv2.PtrString(name),
		Description: identityproviderv2.PtrString(description),
		Issuer:      identityproviderv2.PtrString(issuer),
		JwksUri:     identityproviderv2.PtrString(jwksuri),
	}
	resp, httpResp, err := c.V2Client.CreateIdentityProvider(createIdentityProvider)
	if err != nil {
		return errors.CatchServiceNameInUseError(err, httpResp, name)
	}

	describeIdentityProvider := &identityProvider{Id: *resp.Id, DisplayName: *resp.DisplayName, Description: *resp.Description, Issuer: *resp.Issuer, JwksUri: *resp.JwksUri}

	return output.DescribeObject(cmd, describeIdentityProvider, providerListFields, providerHumanLabelMap, providerStructuredLabelMap)
}
