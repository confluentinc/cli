package streamshare

import (
	"fmt"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
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
