package iam

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *certificatePoolCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List certificate pools.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	c.AddProviderFlag(cmd)
	cobra.CheckErr(cmd.MarkFlagRequired("provider"))
	return cmd
}

func (c *certificatePoolCommand) list(cmd *cobra.Command, _ []string) error {
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		return nil
	}

	certificatePools, err := c.V2Client.ListCertificatePool(provider)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, certificatePool := range certificatePools {
		list.Add(&certificatePoolOut{
			Id:                 certificatePool.GetId(),
			Name:               certificatePool.GetDisplayName(),
			Description:        certificatePool.GetDescription(),
			ExternalIdentifier: certificatePool.GetExternalIdentifier(),
			Filter:             certificatePool.GetFilter(),
		})
	}
	return list.Print()
}
