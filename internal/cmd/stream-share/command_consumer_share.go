package streamshare

import (
	"time"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

var (
	consumerShareListFields           = []string{"Id", "ProviderUserName", "Status", "SharedResourceId", "InviteExpiresAt"}
	consumerShareListHumanLabels      = []string{"ID", "Provider Name", "Status", "Shared Resource ID", "Invite Expires At"}
	consumerShareListStructuredLabels = []string{"id", "provider_name", "status", "shared_resource_id", "invite_expires_at"}
)

var (
	consumerHumanLabelMap = map[string]string{
		"Id":               "ID",
		"ProviderUserName": "Provider Name",
		"Status":           "Status",
		"SharedResourceId": "Shared Resource ID",
		"InviteExpiresAt":  "Invite Expires At",
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

type consumerShareCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newConsumerShareCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share",
		Short: "Manage consumer shares.",
	}

	c := &consumerShareCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newListCommand())

	return c.Command
}

func (s *consumerShareCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := s.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return s.autocompleteConsumerShares()
}

func (s *consumerShareCommand) autocompleteConsumerShares() []string {
	consumerShares, err := s.V2Client.ListConsumerShares("")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(consumerShares))
	for i, share := range consumerShares {
		suggestions[i] = *share.Id
	}
	return suggestions
}

func (s *consumerShareCommand) buildConsumerShare(share cdxv1.CdxV1ConsumerShare) *consumerShare {
	sharedResource := share.GetSharedResource()
	return &consumerShare{
		Id:               share.GetId(),
		ProviderUserName: share.GetProviderUserName(),
		Status:           share.GetStatus(),
		SharedResourceId: sharedResource.GetId(),
		InviteExpiresAt:  share.GetInviteExpiresAt(),
	}
}
