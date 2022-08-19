package streamshare

import (
	"time"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
)

var (
	consumerShareListFields           = []string{"Id", "ProviderName", "Status", "SharedResourceId", "InviteExpiresAt"}
	consumerShareListHumanLabels      = []string{"ID", "Provider Name", "Status", "Shared Resource ID", "Invite Expires At"}
	consumerShareListStructuredLabels = []string{"id", "provider_name", "status", "shared_resource_id", "invite_expires_at"}
)

var (
	consumerHumanLabelMap = map[string]string{
		"Id":               "ID",
		"ProviderName":     "Provider Name",
		"Status":           "Status",
		"SharedResourceId": "Shared Resource ID",
		"InviteExpiresAt":  "Invite Expires At",
	}
	consumerStructuredLabelMap = map[string]string{
		"Id":               "id",
		"ProviderName":     "provider_name",
		"Status":           "status",
		"SharedResourceId": "shared_resource_id",
		"InviteExpiresAt":  "invite_expires_at",
	}
)

type consumerShare struct {
	Id               string
	ProviderName     string
	Status           string
	SharedResourceId string
	InviteExpiresAt  time.Time
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

func (c *command) buildConsumerShare(share cdxv1.CdxV1ConsumerShare) *consumerShare {
	sharedResource := share.GetSharedResource()
	return &consumerShare{
		Id:               share.GetId(),
		ProviderName:     share.GetProviderUserName(),
		Status:           share.GetStatus(),
		SharedResourceId: sharedResource.GetId(),
		InviteExpiresAt:  share.GetInviteExpiresAt(),
	}
}
