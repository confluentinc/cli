package streamshare

import (
	"time"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
)

type shareOut struct {
	Id               string    `human:"ID" serialized:"id"`
	ProviderName     string    `human:"Provider Name" serialized:"provider_name"`
	Status           string    `human:"Status" serialized:"status"`
	SharedResourceId string    `human:"Shared Resource ID" serialized:"shared_resource_id"`
	InviteExpiresAt  time.Time `human:"Invite Expires At" serialized:"invite_expires_at"`
}

func (c *command) newConsumerShareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share",
		Short: "Manage consumer shares.",
	}

	cmd.AddCommand(c.newConsumerShareDeleteCommand())
	cmd.AddCommand(c.newConsumerShareDescribeCommand())
	cmd.AddCommand(c.newConsumerShareListCommand())

	return cmd
}

func (c *command) validConsumerShareArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteConsumerShares()
}

func (c *command) autocompleteConsumerShares() []string {
	consumerShares, err := c.V2Client.ListConsumerShares("")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(consumerShares))
	for i, share := range consumerShares {
		suggestions[i] = *share.Id
	}
	return suggestions
}

func (c *command) buildConsumerShare(share cdxv1.CdxV1ConsumerShare) *shareOut {
	sharedResource := share.GetSharedResource()
	return &shareOut{
		Id:               share.GetId(),
		ProviderName:     share.GetProviderUserName(),
		Status:           share.GetStatus(),
		SharedResourceId: sharedResource.GetId(),
		InviteExpiresAt:  share.GetInviteExpiresAt(),
	}
}
