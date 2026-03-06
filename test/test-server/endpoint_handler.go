package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	endpointv1 "github.com/confluentinc/ccloud-sdk-go-v2/endpoint/v1"
)

// Handler for: "/endpoint/v1/endpoints"
func handleEndpoints(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleEndpointList(t)(w, r)
		}
	}
}

func handleEndpointList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		service := r.URL.Query().Get("service")
		cloud := r.URL.Query().Get("cloud")
		region := r.URL.Query().Get("region")
		environment := r.URL.Query().Get("environment")
		resource := r.URL.Query().Get("resource")
		isPrivate := r.URL.Query().Get("is_private")

		// Build mock response based on filters
		endpoints := []endpointv1.EndpointV1Endpoint{}

		// Helper function to check if endpoint should be included based on privacy filter
		shouldIncludeByPrivacy := func(endpointIsPrivate bool) bool {
			if isPrivate == "" {
				return true // No filter, include all
			}
			if isPrivate == "true" && endpointIsPrivate {
				return true // Filter for private, endpoint is private
			}
			if isPrivate == "false" && !endpointIsPrivate {
				return true // Filter for public, endpoint is public
			}
			return false
		}

		// Mock endpoint 1: AWS, private, KAFKA
		if (cloud == "" || cloud == "AWS") &&
			(region == "" || region == "us-west-2") &&
			(service == "KAFKA") &&
			(resource == "" || resource == "lkc-abc123") &&
			shouldIncludeByPrivacy(true) {
			endpoint1 := createMockEndpoint(
				"e-12345",
				"AWS",
				"us-west-2",
				"KAFKA",
				true,
				"PRIVATE_LINK",
				"https://lkc-abc123-ap12345.us-west-2.aws.glb.confluent.cloud:443",
				"REST",
				environment,
				"lkc-abc123",
				"gw-12345",
				"ap-12345",
			)
			endpoints = append(endpoints, endpoint1)
		}

		// Mock endpoint 2: AWS, public, KAFKA
		if (cloud == "" || cloud == "AWS") &&
			(region == "" || region == "us-east-1") &&
			(service == "KAFKA") &&
			(resource == "" || resource == "lkc-xyz789") &&
			shouldIncludeByPrivacy(false) {
			endpoint2 := createMockEndpoint(
				"e-67890",
				"AWS",
				"us-east-1",
				"KAFKA",
				false,
				"PUBLIC",
				"https://pkc-xyz789.us-east-1.aws.confluent.cloud:443",
				"REST",
				environment,
				"lkc-xyz789",
				"",
				"",
			)
			endpoints = append(endpoints, endpoint2)
		}

		// Mock endpoint 3: GCP, private, SCHEMA_REGISTRY
		if (cloud == "" || cloud == "GCP") &&
			(region == "" || region == "us-central1") &&
			(service == "SCHEMA_REGISTRY") &&
			shouldIncludeByPrivacy(true) {
			endpoint3 := createMockEndpoint(
				"e-abc456",
				"GCP",
				"us-central1",
				"SCHEMA_REGISTRY",
				true,
				"PRIVATE_SERVICE_CONNECT",
				"https://psrc-abc456.us-central1.gcp.confluent.cloud",
				"REST",
				environment,
				"lsrc-def789",
				"gw-67890",
				"ap-34567",
			)
			endpoints = append(endpoints, endpoint3)
		}

		// Build response
		response := endpointv1.EndpointV1EndpointList{
			ApiVersion: "endpoint/v1",
			Kind:       "EndpointList",
			Metadata: endpointv1.ListMeta{
				First: *endpointv1.NewNullableString(endpointv1.PtrString("https://api.confluent.cloud/endpoint/v1/endpoints")),
				Last:  *endpointv1.NewNullableString(endpointv1.PtrString("https://api.confluent.cloud/endpoint/v1/endpoints")),
			},
			Data: endpoints,
		}

		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}
}

func createMockEndpoint(id, cloud, region, service string, isPrivate bool, connectionType, endpoint, endpointType, environment, resource, gateway, accessPoint string) endpointv1.EndpointV1Endpoint {
	ep := endpointv1.EndpointV1Endpoint{
		ApiVersion: endpointv1.PtrString("endpoint/v1"),
		Kind:       endpointv1.PtrString("Endpoint"),
		Id:         endpointv1.PtrString(id),
		Cloud:      endpointv1.PtrString(cloud),
		Region:     endpointv1.PtrString(region),
		Service:    endpointv1.PtrString(service),
		IsPrivate:  endpointv1.PtrBool(isPrivate),
		Endpoint:   endpointv1.PtrString(endpoint),
	}

	if connectionType != "" {
		ep.ConnectionType = endpointv1.PtrString(connectionType)
	}
	if endpointType != "" {
		ep.EndpointType = endpointv1.PtrString(endpointType)
	}
	if environment != "" {
		ep.Environment = &endpointv1.ObjectReference{
			Id: environment,
		}
	}
	if resource != "" {
		ep.Resource = &endpointv1.TypedEnvScopedObjectReference{
			Id: resource,
		}
	}
	if gateway != "" {
		ep.Gateway = &endpointv1.ObjectReference{
			Id: gateway,
		}
	}
	if accessPoint != "" {
		ep.AccessPoint = &endpointv1.ObjectReference{
			Id: accessPoint,
		}
	}

	return ep
}
