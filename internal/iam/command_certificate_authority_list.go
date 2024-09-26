package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *certificateAuthorityCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List certificate authorities.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *certificateAuthorityCommand) list(cmd *cobra.Command, _ []string) error {
	certificateAuthorities, err := c.V2Client.ListCertificateAuthorities()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, certificateAuthority := range certificateAuthorities {
		list.Add(&certificateAuthorityOut{
			Id:                       certificateAuthority.GetId(),
			Name:                     certificateAuthority.GetDisplayName(),
			Description:              certificateAuthority.GetDescription(),
			Fingerprints:             certificateAuthority.GetFingerprints(),
			ExpirationDates:          certificateAuthority.GetExpirationDates(),
			SerialNumbers:            certificateAuthority.GetSerialNumbers(),
			CertificateChainFilename: certificateAuthority.GetCertificateChainFilename(),
			State:                    certificateAuthority.GetState(),
		})
	}
	return list.Print()
}
