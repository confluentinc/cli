package iam

import (
	"github.com/spf13/cobra"

	certificateauthorityv2 "github.com/confluentinc/ccloud-sdk-go-v2/certificate-authority/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *certificateAuthorityCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a certificate authority.",
		Args:              cobra.ExactArgs(1),
		RunE:              c.update,
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the certificate chain for certificate authority "op-123456" using the certificate chain stored in the "CERTIFICATE_CHAIN" environment variable:`,
				Code: "confluent iam certificate-authority update op-123456 --certificate-chain $CERTIFICATE_CHAIN --certificate-chain-filename certificate.pem",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the certificate authority.")
	cmd.Flags().String("description", "", "Description of the certificate authority.")
	cmd.Flags().String("certificate-chain", "", "A base64 encoded string containing the signing certificate chain.")
	cmd.Flags().String("certificate-chain-filename", "", "The name of the certificate file.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsRequiredTogether("certificate-chain", "certificate-chain-filename")

	return cmd
}

func (c *certificateAuthorityCommand) update(cmd *cobra.Command, args []string) error {
	// The update sends a PUT request, so we also need to send unchanged fields
	currentCertificateAuthority, err := c.V2Client.GetCertificateAuthority(args[0])
	if err != nil {
		return err
	}

	update := certificateauthorityv2.IamV2UpdateCertRequest{
		Id:                       certificateauthorityv2.PtrString(args[0]),
		DisplayName:              currentCertificateAuthority.DisplayName,
		Description:              currentCertificateAuthority.Description,
		CertificateChainFilename: currentCertificateAuthority.CertificateChainFilename,
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
	if cmd.Flags().Changed("certificate-chain") && cmd.Flags().Changed("certificate-chain-filename") {
		certificateChain, err := cmd.Flags().GetString("certificate-chain")
		if err != nil {
			return err
		}
		update.CertificateChain = certificateauthorityv2.PtrString(certificateChain)

		certificateChainFilename, err := cmd.Flags().GetString("certificate-chain-filename")
		if err != nil {
			return err
		}
		update.CertificateChainFilename = certificateauthorityv2.PtrString(certificateChainFilename)
	}

	certificateAuthority, err := c.V2Client.UpdateCertificateAuthority(update)
	if err != nil {
		return err
	}
	return printCertificateAuthority(cmd, certificateAuthority)
}
