package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	providerHumanLabelMap = map[string]string{
		"Id":        "ID",
		"IssuerUri": "Issuer URI",
		"JwksUri":   "JWKS URI",
	}
	providerStructuredLabelMap = map[string]string{
		"Id":          "id",
		"Description": "description",
		"IssuerUri":   "issuer_uri",
		"JwksUri":     "jwks_uri",
	}
)

func (c identityProviderCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe an identity provider.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c identityProviderCommand) describe(cmd *cobra.Command, args []string) error {
	identityProviderProfile, httpResp, err := c.V2Client.GetIdentityProvider(args[0])
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	describeIdentityProvider := &identityProvider{
		Id:        *identityProviderProfile.Id,
		Name:      *identityProviderProfile.DisplayName,
		IssuerUri: *identityProviderProfile.Issuer,
		JwksUri:   *identityProviderProfile.JwksUri,
	}
	if identityProviderProfile.Description != nil {
		describeIdentityProvider.Description = *identityProviderProfile.Description
	}

	return output.DescribeObject(cmd, describeIdentityProvider, providerListFields, providerHumanLabelMap, providerStructuredLabelMap)
}
