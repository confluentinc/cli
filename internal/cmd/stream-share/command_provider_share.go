package streamshare

import (
	streamsharev1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/cdx/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
	"net/url"
	"time"
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
		"ConsumerUserName":         "consumer_user_name",
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
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	s := &providerShareCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	s.AddCommand(s.newDeleteCommand())
	s.AddCommand(s.newDescribeCommand())
	s.AddCommand(s.newListCommand())

	return s.Command
}

func (s *providerShareCommand) list(cmd *cobra.Command, _ []string) error {
	sharedResource, err := cmd.Flags().GetString("shared_resource")
	if err != nil {
		return err
	}

	var sharesList []streamsharev1.CdxV1ProviderShare
	var token string

	for {
		listResult, _, err := s.V2Client.ListProviderShares(sharedResource, token)
		if err != nil {
			return err
		}
		sharesList = append(sharesList, listResult.Data...)

		token = ""
		if md, ok := listResult.GetMetadataOk(); ok && md.GetNext() != "" {
			parsed, err := url.Parse(*md.Next.Get())
			if err != nil {
				return err
			}
			token = parsed.Query().Get("page_token")
		}

		if token == "" {
			break
		}
	}

	outputWriter, err := output.NewListOutputWriter(cmd, providerShareListFields, providerShareListHumanLabels,
		providerShareListStructuredLabels)
	if err != nil {
		return err
	}

	for _, share := range sharesList {
		element := s.buildProviderShare(share)

		outputWriter.AddElement(element)
	}

	return outputWriter.Out()
}

func (s *providerShareCommand) buildProviderShare(share streamsharev1.CdxV1ProviderShare) *providerShare {
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

func (s *providerShareCommand) describe(cmd *cobra.Command, args []string) error {
	shareId := args[0]

	provideShare, _, err := s.V2Client.DescribeProvideShare(shareId)
	if err != nil {
		return err
	}

	return output.DescribeObject(cmd, s.buildProviderShare(provideShare), providerShareListFields, humanLabelMap,
		structuredLabelMap)
}

func (s *providerShareCommand) delete(cmd *cobra.Command, args []string) error {
	shareId := args[0]

	_, err := s.V2Client.DeleteProviderShare(shareId)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.DeletedProviderShareMsg, shareId)
	return nil
}
