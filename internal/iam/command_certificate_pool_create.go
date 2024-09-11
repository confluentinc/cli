package iam

import (
	"github.com/spf13/cobra"

	certificateauthorityv2 "github.com/confluentinc/ccloud-sdk-go-v2/certificate-authority/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *certificatePoolCommand) newCreateCommand() *cobra.Command {
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

	createCertificatePool := certificateauthorityv2.IamV2CertificateIdentityPool{
		DisplayName:        certificateauthorityv2.PtrString(args[0]),
		Description:        certificateauthorityv2.PtrString(description),
		Filter:             certificateauthorityv2.PtrString(filter),
		ExternalIdentifier: certificateauthorityv2.PtrString(externalIdentifier),
	}
	certificatePool, err := c.V2Client.CreateCertificatePool(createCertificatePool, provider)
	if err != nil {
		return err
	}
	return printCertificatePool(cmd, certificatePool)
}
