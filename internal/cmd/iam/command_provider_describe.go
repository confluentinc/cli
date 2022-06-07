package iam

import (
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	providerHumanLabelMap = map[string]string{
		"Id":          "ID",
		"DisplayName": "Display Name",
		"Description": "Description",
		"Issuer":      "Issuer",
		"JwksUri":     "JWKS URI",
	}
	providerStructuredLabelMap = map[string]string{
		"Id":          "id",
		"DisplayName": "display_name",
		"Description": "description",
		"Issuer":      "issuer",
		"JwksUri":     "jwks_uri",
	}
)

func (c identityProviderCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe an identity provider.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c identityProviderCommand) describe(cmd *cobra.Command, args []string) error {
	if resource.LookupType(args[0]) != resource.IdentityProvider {
		return errors.New(errors.BadIdentityProviderIDErrorMsg)
	}

	identityProviderProfile, _, err := c.V2Client.GetIdentityProvider(args[0])
	if err != nil {
		return err
	}

	return output.DescribeObject(cmd, &identityProvider{
		Id:          *identityProviderProfile.Id,
		DisplayName: *identityProviderProfile.DisplayName,
		Description: *identityProviderProfile.Description,
		Issuer:      *identityProviderProfile.Issuer,
		JwksUri:     *identityProviderProfile.JwksUri,
	}, providerListFields, providerHumanLabelMap, providerStructuredLabelMap)
}
