package streamshare

import (
	"fmt"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
)

type privateLinkNetworkDetails struct {
	networkKind         string
	privateLinkDataType string
	privateLinkData     string
}

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
		details.privateLinkData = fmt.Sprintf("%v", cloud.CdxV1AzureNetwork.GetPrivateLinkServiceAliases())
	} else if cloud.CdxV1GcpNetwork != nil {
		details.networkKind = cloud.CdxV1GcpNetwork.Kind
		details.privateLinkDataType = "Private Service Connect Service Attachments"
		details.privateLinkData = fmt.Sprintf("%v", cloud.CdxV1GcpNetwork.GetPrivateServiceConnectServiceAttachments())
	}
	return &details
}

func mapSubdomainsToList(m map[string]string) []string {
	subdomains := make([]string, len(m))
	i := 0
	for k, v := range m {
		subdomains[i] = fmt.Sprintf(`%s="%s"`, k, v)
		i++
	}
	return subdomains
}

func confirmOptOut() (bool, error) {
	f := form.New(
		form.Field{
			ID: "confirmation",
			Prompt: "Are you sure you want to disable Stream Sharing for your organization? " +
				"Existing shares in your organization will not be accessible if Stream Sharing is disabled.",
			IsYesOrNo: true,
		},
	)
	if err := f.Prompt(form.NewPrompt()); err != nil {
		return false, errors.New(errors.FailedToReadInputErrorMsg)
	}
	return f.Responses["confirmation"].(bool), nil
}
