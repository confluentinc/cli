package iam

import (
	"time"

	"github.com/spf13/cobra"

	certificateauthorityv2 "github.com/confluentinc/ccloud-sdk-go-v2/certificate-authority/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type certificateAuthorityCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type certificateAuthorityOut struct {
	Id                       string      `human:"ID" serialized:"id"`
	Name                     string      `human:"Name" serialized:"name"`
	Description              string      `human:"Description" serialized:"description"`
	Fingerprints             []string    `human:"Fingerprints" serialized:"fingerprints"`
	ExpirationDates          []time.Time `human:"Expiration Dates" serialized:"expiration_dates"`
	SerialNumbers            []string    `human:"Serial Numbers" serialized:"serial_numbers"`
	CertificateChainFilename string      `human:"Certificate Chain Filename" serialized:"certificate_chain_filename"`
	CrlSource                string      `human:"CRL Source,omitempty" serialized:"crl_source,omitempty"`
	CrlUrl                   string      `human:"CRL URL,omitempty" serialized:"crl_url,omitempty"`
	CrlUpdatedAt             *time.Time  `human:"CRL Updated At,omitempty" serialized:"crl_updated_at,omitempty"`
}

func newCertificateAuthorityCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "certificate-authority",
		Short:       "Manage certificate authorities.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &certificateAuthorityCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printCertificateAuthority(cmd *cobra.Command, certificateAuthority certificateauthorityv2.IamV2CertificateAuthority) error {
	table := output.NewTable(cmd)
	table.Add(&certificateAuthorityOut{
		Id:                       certificateAuthority.GetId(),
		Name:                     certificateAuthority.GetDisplayName(),
		Description:              certificateAuthority.GetDescription(),
		Fingerprints:             certificateAuthority.GetFingerprints(),
		ExpirationDates:          certificateAuthority.GetExpirationDates(),
		SerialNumbers:            certificateAuthority.GetSerialNumbers(),
		CertificateChainFilename: certificateAuthority.GetCertificateChainFilename(),
		CrlSource:                certificateAuthority.GetCrlSource(),
		CrlUrl:                   certificateAuthority.GetCrlUrl(),
		CrlUpdatedAt:             certificateAuthority.CrlUpdatedAt,
	})
	return table.Print()
}

func (c *certificateAuthorityCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *certificateAuthorityCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteCertificateAuthorities(c.V2Client)
}
