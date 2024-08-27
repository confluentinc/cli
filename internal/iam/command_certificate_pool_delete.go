package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *certificatePoolCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more certificate pools.",
		Args:              cobra.MinimumNArgs(1),
		RunE:              c.delete,
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete certificate pool "pool-123":`,
				Code: "confluent iam certificate-pool delete pool-123 --provider provider-123",
			},
		),
	}

	c.AddProviderFlag(cmd)
	pcmd.AddForceFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("provider"))

	return cmd
}

func (c *certificatePoolCommand) delete(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return nil
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetCertificatePool(id, provider)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.CertificatePool); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteCertificatePool(id, provider)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.CertificatePool)
	return err
}
