package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *certificatePoolCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a certificate pool.",
		Args:              cobra.ExactArgs(1),
		RunE:              c.describe,
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe a certificate pool with ID "pool-123".`,
				Code: "confluent iam certificate-pool describe pool-123 --provider provider-123",
			},
		),
	}

	c.AddProviderFlag(cmd)
	cobra.CheckErr(cmd.MarkFlagRequired("provider"))
	return cmd
}

func (c *certificatePoolCommand) describe(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return nil
	}
	certificatePool, err := c.V2Client.GetCertificatePool(args[0], provider)
	if err != nil {
		return err
	}
	return printCertificatePool(cmd, certificatePool)
}
