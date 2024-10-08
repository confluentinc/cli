package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newConsumerShareListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List consumer shares.",
		Args:  cobra.NoArgs,
		RunE:  c.listConsumerShares,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List consumer shares for shared resource "sr-12345":`,
				Code: "confluent stream-share consumer share list --shared-resource sr-12345",
			},
		),
	}

	cmd.Flags().String("shared-resource", "", "Filter the results by a shared resource.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) listConsumerShares(cmd *cobra.Command, _ []string) error {
	sharedResource, err := cmd.Flags().GetString("shared-resource")
	if err != nil {
		return err
	}

	consumerShares, err := c.V2Client.ListConsumerShares(sharedResource)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, share := range consumerShares {
		list.Add(&consumerShareOut{
			Id:                       share.GetId(),
			ProviderName:             share.GetProviderUserName(),
			ProviderOrganizationName: share.GetProviderOrganizationName(),
			Status:                   share.Status.GetPhase(),
			InviteExpiresAt:          share.GetInviteExpiresAt(),
		})
	}
	list.Filter([]string{"Id", "ProviderName", "ProviderOrganizationName", "Status", "InviteExpiresAt"})
	return list.Print()
}
