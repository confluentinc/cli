package streamshare

import (
	"fmt"
	"os"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/spf13/cobra"
)

func getPrivateLinkNetworkDetails(network cdxv1.CdxV1Network) *privateLinkNetworkDetails {
	cloud := network.GetCloud()
	var details privateLinkNetworkDetails
	if cloud.CdxV1AwsNetwork != nil {
		details.networkKind = cloud.CdxV1AwsNetwork.Kind
		details.privateLinkDataType = "Private Link Endpoint Service"
		details.privateLinkData = cloud.CdxV1AwsNetwork.GetPrivateLinkEndpointService()
	} else if cloud.CdxV1AzureNetwork != nil {
		details.networkKind = cloud.CdxV1AzureNetwork.Kind
		details.privateLinkDataType = "Private Link Service Aliases"
		details.privateLinkData = cloud.CdxV1AzureNetwork.GetPrivateLinkServiceAliases()
	} else if cloud.CdxV1GcpNetwork != nil {
		details.networkKind = cloud.CdxV1GcpNetwork.Kind
		details.privateLinkDataType = "Private Service Connect Service Attachments"
		details.privateLinkData = cloud.CdxV1GcpNetwork.GetPrivateServiceConnectServiceAttachments()
	}
	return &details
}

func combineMaps(m1, m2 map[string]string) map[string]string {
	for k, v := range m2 {
		m1[k] = v
	}
	return m1
}

func mapSubdomainsToList(m map[string]string) []string {
	var subdomains []string
	for k, v := range m {
		subdomains = append(subdomains, fmt.Sprintf(`%s="%s"`, k, v))
	}

	return subdomains
}

func confirmOptOut(cmd *cobra.Command) (bool, error) {
	f := form.New(
		form.Field{
			ID: "confirmation",
			Prompt: "Are you sure you want to disable Stream Sharing for your organization? " +
				"Existing shares in your organization will not be accessible if Stream Sharing is disabled.",
			IsYesOrNo: true,
		},
	)
	if err := f.Prompt(cmd, form.NewPrompt(os.Stdin)); err != nil {
		return false, errors.New(errors.FailedToReadOptOutConfirmationErrorMsg)
	}
	return f.Responses["confirmation"].(bool), nil
}
