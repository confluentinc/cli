package streamshare

import (
	"time"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
)

var (
	consumerShareListFields           = []string{"Id", "ProviderUserName", "Status", "SharedResourceId", "InviteExpiresAt"}
	consumerShareListHumanLabels      = []string{"ID", "Provider Name", "Status", "Shared Resource ID", "Invite Expiration"}
	consumerShareListStructuredLabels = []string{"id", "provider_name", "status", "shared_resource_id", "invite_expires_at"}
)

var (
	consumerHumanLabelMap = map[string]string{
		"Id":               "ID",
		"ProviderUserName": "Provider Name",
		"Status":           "Status",
		"SharedResourceId": "Shared Resource ID",
		"InviteExpiresAt":  "Invite Expiration",
	}
	consumerStructuredLabelMap = map[string]string{
		"Id":               "id",
		"ProviderUserName": "provider_name",
		"Status":           "status",
		"SharedResourceId": "shared_resource_id",
		"InviteExpiresAt":  "invite_expires_at",
	}
)

type consumerShare struct {
	Id               string
	ProviderUserName string
	Status           string
	SharedResourceId string
	InviteExpiresAt  time.Time
}

func (c *command) newConsumerShareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share",
		Short: "Manage consumer shares.",
	}

	cmd.AddCommand(c.newDeleteConsumerShareCommand())
	cmd.AddCommand(c.newDescribeConsumerShareCommand())
	cmd.AddCommand(c.newListConsumerSharesCommand())

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

func (c *command) buildConsumerShare(share cdxv1.CdxV1ConsumerShare) *consumerShare {
	sharedResource := share.GetSharedResource()
	return &consumerShare{
		Id:               share.GetId(),
		ProviderUserName: share.GetProviderUserName(),
		Status:           share.GetStatus(),
		SharedResourceId: sharedResource.GetId(),
		InviteExpiresAt:  share.GetInviteExpiresAt(),
	}
}
