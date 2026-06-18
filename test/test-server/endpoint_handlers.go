package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	endpointv1 "github.com/confluentinc/ccloud-sdk-go-v2/endpoint/v1"
)

// handleEndpointV1Endpoints mocks the new Endpoints API.
// For each supported (cloud, region) pair it returns the same set of FLINK endpoints
// the legacy three-source aggregation produced, so the list-*.golden fixtures
// remain unchanged. The handler also emits at least one LANGUAGE_SERVICE
// (`flinkpls.*`) endpoint to exercise the CLI's REST-only filter.
func handleEndpointV1Endpoints(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// The real API returns cloud values lowercased; the CLI uppercases them for
		// display. We match the real API here so the test exercises that conversion.
		cloud := r.URL.Query().Get("cloud")
		region := r.URL.Query().Get("region")

		var endpoints []endpointv1.EndpointV1Endpoint
		switch {
		case cloud == "AWS" && region == "eu-west-1":
			endpoints = []endpointv1.EndpointV1Endpoint{
				newFlinkRestEndpoint("aws", region, TestFlinkGatewayUrl.String(), false),
				newFlinkRestEndpoint("aws", region, TestFlinkGatewayUrlPrivate.String(), true),
				newFlinkRestEndpoint("aws", region, "https://flink-n-abcde6.eu-west-1.aws.confluent.cloud", true),
				// LANGUAGE_SERVICE endpoint must be filtered out by the CLI.
				newFlinkLanguageServiceEndpoint("aws", region, "https://flinkpls.eu-west-1.aws.confluent.cloud", false),
			}
		case cloud == "AZURE" && region == "centralus":
			endpoints = []endpointv1.EndpointV1Endpoint{
				newFlinkRestEndpoint("azure", region, TestFlinkGatewayUrl.String(), false),
			}
		case cloud == "AZURE" && region == "eastus2":
			endpoints = []endpointv1.EndpointV1Endpoint{
				newFlinkRestEndpoint("azure", region, TestFlinkGatewayUrl.String(), false),
				newFlinkRestEndpoint("azure", region, "https://flink-n-abcde2.eastus.azure.confluent.cloud", true),
				newFlinkRestEndpoint("azure", region, "https://flink-n-abcde7.eastus.azure.confluent.cloud", true),
			}
		case cloud == "GCP" && region == "europe-west3-a":
			endpoints = []endpointv1.EndpointV1Endpoint{
				newFlinkRestEndpoint("gcp", region, TestFlinkGatewayUrl.String(), false),
				newFlinkRestEndpoint("gcp", region, TestFlinkGatewayUrlPrivate.String(), true),
			}
		case cloud == "AZURE" && region == "italynorth":
			// Multi-PLATT (PrivateLink Gateway) shape: an access-point URL on the GLB
			// domain that the legacy URL-template aggregation could not have constructed.
			// Mirrors a real production env captured during FCP-4223 verification.
			endpoints = []endpointv1.EndpointV1Endpoint{
				newFlinkRestEndpoint("azure", region, "https://flink.italynorth.azure.confluent.cloud", false),
				newFlinkRestEndpoint("azure", region, "https://flink-ap4jnpj9.italynorth.azure.accesspoint.glb.confluent.cloud", true),
				// LANGUAGE_SERVICE row that the CLI must filter out.
				newFlinkLanguageServiceEndpoint("azure", region, "https://flinkpls.italynorth.azure.confluent.cloud", false),
			}
		}

		list := &endpointv1.EndpointV1EndpointList{Data: endpoints}
		err := json.NewEncoder(w).Encode(list)
		require.NoError(t, err)
	}
}

func newFlinkRestEndpoint(cloud, region, url string, isPrivate bool) endpointv1.EndpointV1Endpoint {
	return newFlinkEndpoint(cloud, region, url, isPrivate, "REST")
}

func newFlinkLanguageServiceEndpoint(cloud, region, url string, isPrivate bool) endpointv1.EndpointV1Endpoint {
	return newFlinkEndpoint(cloud, region, url, isPrivate, "LANGUAGE_SERVICE")
}

func newFlinkEndpoint(cloud, region, url string, isPrivate bool, endpointType string) endpointv1.EndpointV1Endpoint {
	return endpointv1.EndpointV1Endpoint{
		Cloud:        endpointv1.PtrString(cloud),
		Region:       endpointv1.PtrString(region),
		Service:      endpointv1.PtrString("FLINK"),
		Endpoint:     endpointv1.PtrString(url),
		IsPrivate:    endpointv1.PtrBool(isPrivate),
		EndpointType: endpointv1.PtrString(endpointType),
	}
}
