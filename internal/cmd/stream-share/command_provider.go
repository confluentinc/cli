package streamshare

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	providerShareListFields = []string{"Id", "ConsumerName", "ConsumerOrganizationName", "ProviderName",
		"Status", "DeliveryMethod", "ServiceAccountId", "SharedResourceId", "InvitedAt", "RedeemedAt", "InviteExpiresAt"}
	providerShareListHumanLabels = []string{"ID", "Consumer Name", "Consumer Organization Name", "Provider Name",
		"Status", "Delivery Method", "Service Account ID", "Shared Resource ID", "Invited At", "Redeemed At", "Invite Expires At"}
	providerShareListStructuredLabels = []string{"id", "consumer_name", "consumer_organization_name", "provider_name",
		"status", "delivery_method", "service_account_id", "shared_resource_id", "invited_at", "redeemed_at", "invite_expires_at"}
)

type providerShare struct {
	Id                       string
	ConsumerName             string
	ConsumerOrganizationName string
	ProviderName             string
	Status                   string
	DeliveryMethod           string
	ServiceAccountId         string
	SharedResourceId         string
	RedeemedAt               string
	InvitedAt                time.Time
	InviteExpiresAt          time.Time
}

var (
	providerHumanLabelMap = map[string]string{
		"Id":                       "ID",
		"ConsumerName":             "Consumer Name",
		"ConsumerOrganizationName": "Consumer Organization Name",
		"ProviderName":             "Provider Name",
		"Status":                   "Status",
		"DeliveryMethod":           "Delivery Method",
		"ServiceAccount":           "Service Account",
		"SharedResource":           "Shared Resource",
		"RedeemedAt":               "Redeemed At",
		"InvitedAt":                "Invited At",
		"InviteExpiresAt":          "Invite Expires At",
	}
	providerStructuredLabelMap = map[string]string{
		"Id":                       "id",
		"ConsumerName":             "consumer_name",
		"ConsumerOrganizationName": "consumer_organization_name",
		"ProviderName":             "provider_name",
		"Status":                   "status",
		"DeliveryMethod":           "delivery_method",
		"ServiceAccountId":         "service_account_id",
		"SharedResourceId":         "shared_resource_id",
		"RedeemedAt":               "redeemed_at",
		"InvitedAt":                "invited_at",
		"InviteExpiresAt":          "invite_expires_at",
	}
)

func (c *command) newProviderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage provider actions.",
	}

	cmd.AddCommand(c.newInviteCommand())
	cmd.AddCommand(c.newProviderShareCommand())

	return cmd
}
