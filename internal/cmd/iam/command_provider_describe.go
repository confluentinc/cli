package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	providerHumanLabelMap = map[string]string{
		"Id":          "ID",
		"DisplayName": "Display Name",
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
	identityProviderProfile, _, err := c.V2Client.GetIdentityProvider(args[0])
	if err != nil {
		return err
	}

	describeIdentityProvider := &identityProvider{
		Id:          *identityProviderProfile.Id,
		DisplayName: *identityProviderProfile.DisplayName,
		Description: *identityProviderProfile.Description,
		Issuer:      *identityProviderProfile.Issuer,
		JwksUri:     *identityProviderProfile.JwksUri,
	}

	return output.DescribeObject(cmd, describeIdentityProvider, providerListFields, providerHumanLabelMap, providerStructuredLabelMap)
}
