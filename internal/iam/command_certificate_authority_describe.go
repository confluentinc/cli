package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *certificateAuthorityCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a certificate authority.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *certificateAuthorityCommand) describe(cmd *cobra.Command, args []string) error {
	certificateAuthority, err := c.V2Client.GetCertificateAuthority(args[0])
	if err != nil {
		return err
	}

	return printCertificateAuthority(cmd, certificateAuthority)
}
