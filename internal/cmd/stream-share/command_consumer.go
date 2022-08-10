package streamshare

import (
	"time"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
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

type consumerCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newConsumerCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "Manage consumer actions.",
	}

	c := &consumerCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(newConsumerShareCommand(prerunner))
	c.AddCommand(c.newRedeemCommand())

	return c.Command
}

func buildConsumerShare(share cdxv1.CdxV1ConsumerShare) *consumerShare {
	sharedResource := share.GetSharedResource()
	return &consumerShare{
		Id:               share.GetId(),
		ProviderUserName: share.GetProviderUserName(),
		Status:           share.GetStatus(),
		SharedResourceId: sharedResource.GetId(),
		InviteExpiresAt:  share.GetInviteExpiresAt(),
	}
}
