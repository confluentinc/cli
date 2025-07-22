package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *certificateAuthorityCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more certificate authorities.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete certificate authority "op-123456":`,
				Code: "confluent iam certificate-authority delete op-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *certificateAuthorityCommand) delete(cmd *cobra.Command, args []string) error {
	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetCertificateAuthority(id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.CertificateAuthority); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteCertificateAuthority(id)
	}

	_, err := deletion.Delete(cmd, args, deleteFunc, resource.CertificateAuthority)
	return err
}
