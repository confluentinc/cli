package iam

import (
	"github.com/spf13/cobra"

	certificateauthorityv2 "github.com/confluentinc/ccloud-sdk-go-v2/certificate-authority/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *certificatePoolCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a certificate pool.",
		Args:              cobra.ExactArgs(1),
		RunE:              c.update,
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update a certificate pool named "pool-123".`,
				Code: `confluent iam certificate-pool update pool-123 --provider provider-123 --description "update pool"`,
			},
		),
	}

	c.AddProviderFlag(cmd)
	cmd.Flags().String("description", "", "Description of the certificate pool.")
	cmd.Flags().String("name", "", "Name of the certificate pool.")
	pcmd.AddFilterFlag(cmd)
	pcmd.AddExternalIdentifierFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("provider"))

	return cmd
}

func (c *certificatePoolCommand) update(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return err
	}

	// The update sends a PUT request, so we also need to send unchanged fields
	currentCertificatePool, err := c.V2Client.GetCertificatePool(args[0], provider)
	if err != nil {
		return err
	}

	update := certificateauthorityv2.IamV2CertificateIdentityPool{
		Id:                 certificateauthorityv2.PtrString(args[0]),
		DisplayName:        currentCertificatePool.DisplayName,
		Description:        currentCertificatePool.Description,
		Filter:             currentCertificatePool.Filter,
		ExternalIdentifier: currentCertificatePool.ExternalIdentifier,
	}

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		update.DisplayName = certificateauthorityv2.PtrString(name)
	}
	if cmd.Flags().Changed("description") {
		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		update.Description = certificateauthorityv2.PtrString(description)
	}
	if cmd.Flags().Changed("filter") {
		filter, err := cmd.Flags().GetString("filter")
		if err != nil {
			return err
		}
		update.Filter = certificateauthorityv2.PtrString(filter)
	}
	if cmd.Flags().Changed("external-identifier") {
		externalIdentifier, err := cmd.Flags().GetString("external-identifier")
		if err != nil {
			return err
		}
		update.ExternalIdentifier = certificateauthorityv2.PtrString(externalIdentifier)
	}

	certificatePool, err := c.V2Client.UpdateCertificatePool(update, provider)
	if err != nil {
		return err
	}
	return printCertificatePool(cmd, certificatePool)
}
