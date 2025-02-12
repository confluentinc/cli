package iam

import (
	"github.com/spf13/cobra"

	certificateauthorityv2 "github.com/confluentinc/ccloud-sdk-go-v2/certificate-authority/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
)

func (c *certificatePoolCommand) newCreateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "create <name>",
		Short:             "Create a certificate pool.",
		Args:              cobra.ExactArgs(1),
		RunE:              c.create,
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a certificate pool named "pool-123".`,
				Code: `confluent iam certificate-pool create pool-123 --provider provider-123 --description "new description"`,
			},
		),
	}

	c.AddProviderFlag(cmd)
	cmd.Flags().String("description", "", "Description of the certificate pool.")
	pcmd.AddFilterFlag(cmd)
	pcmd.AddExternalIdentifierFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	isResourceOwnerEnabled := cfg.IsTest ||
		(cfg.Context() != nil &&
			featureflags.Manager.BoolVariation("auth.workload_identity.resource_owner.enabled", cfg.Context(), featureflags.GetCcloudLaunchDarklyClient(cfg.Context().PlatformName), true, false))
	if isResourceOwnerEnabled {
		addResourceOwnerFlag(cmd, c.AuthenticatedCLICommand)
	}
	cobra.CheckErr(cmd.MarkFlagRequired("provider"))

	return cmd
}

func (c *certificatePoolCommand) create(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	externalIdentifier, err := cmd.Flags().GetString("external-identifier")
	if err != nil {
		return err
	}
	ldClient := featureflags.GetCcloudLaunchDarklyClient(c.Context.PlatformName)
	isResourceOwnerEnabled := c.Config.IsTest ||
		featureflags.Manager.BoolVariation("auth.workload_identity.resource_owner.enabled", c.Context, ldClient, true, false)
	resourceOwner, err := "", nil
	if isResourceOwnerEnabled {
		resourceOwner, err = cmd.Flags().GetString("resource-owner")
		if err != nil {
			return err
		}
	}

	createCertificatePool := certificateauthorityv2.IamV2CertificateIdentityPool{
		DisplayName:        certificateauthorityv2.PtrString(args[0]),
		Description:        certificateauthorityv2.PtrString(description),
		Filter:             certificateauthorityv2.PtrString(filter),
		ExternalIdentifier: certificateauthorityv2.PtrString(externalIdentifier),
	}
	certificatePool, err := c.V2Client.CreateCertificatePool(createCertificatePool, provider, resourceOwner)
	if err != nil {
		return err
	}
	return printCertificatePool(cmd, certificatePool)
}
