package streamshare

import (
	"time"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/cdx/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/cobra"
)

var (
	providerShareListFields = []string{"Id", "ConsumerUserName", "ConsumerOrganizationName", "ProviderUserName",
		"Status", "DeliveryMethod", "ServiceAccountId", "SharedResourceId", "InvitedAt", "RedeemedAt", "InviteExpiresAt"}
	providerShareListHumanLabels = []string{"ID", "Consumer Name", "Consumer Organization Name", "Provider Name",
		"Status", "Delivery Method", "Service Account ID", "Shared Resource ID", "Invited At", "Redeemed At", "Invite Expires At"}
	providerShareListStructuredLabels = []string{"id", "consumer_user_name", "consumer_organization_name", "provider_user_name",
		"status", "delivery_method", "service_account_id", "shared_resource_id", "invited_at", "redeemed_at", "invite_expires_at"}
)

type providerShare struct {
	Id                       string
	ConsumerUserName         string
	ConsumerOrganizationName string
	ProviderUserName         string
	Status                   string
	DeliveryMethod           string
	ServiceAccountId         string
	SharedResourceId         string
	RedeemedAt               string
	InvitedAt                time.Time
	InviteExpiresAt          time.Time
}

var (
	humanLabelMap = map[string]string{
		"Id":                       "ID",
		"ConsumerUserName":         "Consumer Name",
		"ConsumerOrganizationName": "Consumer Organization Name",
		"ProviderUserName":         "Provider Name",
		"Status":                   "Status",
		"DeliveryMethod":           "Delivery Method",
		"ServiceAccountId":         "Service Account ID",
		"SharedResourceId":         "Shared Resource ID",
		"RedeemedAt":               "Redeemed At",
		"InvitedAt":                "Invited At",
		"InviteExpiresAt":          "Invite Expires At",
	}
	structuredLabelMap = map[string]string{
		"Id":                       "id",
		"ConsumerUserName":         "consumer_name",
		"ConsumerOrganizationName": "consumer_organization_name",
		"ProviderUserName":         "provider_user_name",
		"Status":                   "status",
		"DeliveryMethod":           "delivery_method",
		"ServiceAccountId":         "service_account_id",
		"SharedResourceId":         "shared_resource_id",
		"RedeemedAt":               "redeemed_at",
		"InvitedAt":                "invited_at",
		"InviteExpiresAt":          "invite_expires_at",
	}
)

type providerShareCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newProviderShareCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "share",
		Short:       "Manage provider shares.",
	}

	s := &providerShareCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	s.AddCommand(s.newDeleteCommand())
	s.AddCommand(s.newDescribeCommand())
	s.AddCommand(s.newListCommand())

	return s.Command
}

func (s *providerShareCommand) buildProviderShare(share cdxv1.CdxV1ProviderShare) *providerShare {
	serviceAccount := share.GetServiceAccount()
	sharedResource := share.GetSharedResource()
	element := &providerShare{
		Id:                       share.GetId(),
		ConsumerUserName:         share.GetConsumerUserName(),
		ConsumerOrganizationName: share.GetConsumerOrganizationName(),
		ProviderUserName:         share.GetProviderUserName(),
		Status:                   share.GetStatus(),
		DeliveryMethod:           share.GetDeliveryMethod(),
		ServiceAccountId:         serviceAccount.GetId(),
		SharedResourceId:         sharedResource.GetId(),
		InvitedAt:                share.GetInvitedAt(),
		InviteExpiresAt:          share.GetInviteExpiresAt(),
	}

	if val, ok := share.GetRedeemedAtOk(); ok && !val.IsZero() {
		element.RedeemedAt = val.String()
	}
	return element
}

func (s *providerShareCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := s.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return s.autocompleteProviderShares()
}

func (s *providerShareCommand) autocompleteProviderShares() []string {
	providerShares, err := s.V2Client.ListProviderShares("")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(providerShares))
	for i, share := range providerShares {
		suggestions[i] = *share.Id
	}
	return suggestions
}
