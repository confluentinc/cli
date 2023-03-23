package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
)

func getTestConsumerShare() cdxv1.CdxV1ConsumerShare {
	expiresAt, _ := time.Parse(time.RFC3339, "2022-07-22T22:08:41+00:00")
	return cdxv1.CdxV1ConsumerShare{
		Id:                       stringToPtr("ss-12345"),
		ProviderUserName:         stringToPtr("provider"),
		ProviderOrganizationName: stringPtr("provider org"),
		Status:                   &cdxv1.CdxV1ConsumerShareStatus{Phase: "active"},
		InviteExpiresAt:          &expiresAt,
	}
}

func getTestConsumerSharedResource() cdxv1.CdxV1ConsumerSharedResource {
	return cdxv1.CdxV1ConsumerSharedResource{
		Id: stringToPtr("sr-12345"),
	}
}

func getTestAWSNetwork() *cdxv1.CdxV1AwsNetwork {
	return &cdxv1.CdxV1AwsNetwork{
		Kind:                       "AwsNetwork",
		PrivateLinkEndpointService: stringToPtr("com.amazonaws.vpce.us-west-2.vpce-svc-0000000000"),
	}
}

func getTestProviderShare() cdxv1.CdxV1ProviderShare {
	invitedAt, _ := time.Parse(time.RFC3339, "2022-07-20T22:08:41+00:00")
	redeemedAt, _ := time.Parse(time.RFC3339, "2022-07-21T22:08:41+00:00")
	expiresAt, _ := time.Parse(time.RFC3339, "2022-07-22T22:08:41+00:00")
	return cdxv1.CdxV1ProviderShare{
		Id:                       stringToPtr("ss-12345"),
		ConsumerUserName:         stringToPtr("consumer"),
		ConsumerOrganizationName: stringToPtr("consumer org"),
		Status:                   &cdxv1.CdxV1ProviderShareStatus{Phase: "active"},
		DeliveryMethod:           stringToPtr("email"),
		RedeemedAt:               &redeemedAt,
		InvitedAt:                &invitedAt,
		InviteExpiresAt:          &expiresAt,
	}
}

// Handler for: "/cdx/v1/provider-shares/{id}:resend"
func handleStreamSharingResendInvite(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for: "/cdx/v1/provider-shares"
func handleStreamSharingProviderShares(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			list := cdxv1.CdxV1ProviderShareList{
				Data: []cdxv1.CdxV1ProviderShare{getTestProviderShare()},
			}
			b, err := json.Marshal(&list)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		case http.MethodPost:
			b, err := json.Marshal(getTestProviderShare())
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	}
}

// Handler for: "/cdx/v1/provider-shares/{id}"
func handleStreamSharingProviderShare(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			b, err := json.Marshal(getTestProviderShare())
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/cdx/v1/consumer-shares"
func handleStreamSharingConsumerShares(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list := cdxv1.CdxV1ConsumerShareList{
			Data: []cdxv1.CdxV1ConsumerShare{getTestConsumerShare()},
		}
		b, err := json.Marshal(&list)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for: "/cdx/v1/consumer-shares/{id}"
func handleStreamSharingConsumerShare(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			b, err := json.Marshal(getTestConsumerShare())
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/cdx/v1/shared-tokens:redeem"
func handleStreamSharingRedeemToken(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := cdxv1.CdxV1RedeemTokenResponse{
			Id:                   stringPtr("ss-12345"),
			ApiKey:               stringPtr("00000000000000000000"),
			Secret:               stringPtr("00000000000000000000"),
			KafkaBootstrapUrl:    stringPtr("pkc-00000.us-east1.gcp.confluent.cloud:9092"),
			SchemaRegistryUrl:    stringToPtr("https://psrc-xyz123.us-west-2.aws.cpdev.cloud"),
			SchemaRegistryApiKey: stringPtr("00000000000000000000"),
			SchemaRegistrySecret: stringPtr("00000000000000000000"),
			Resources: &[]cdxv1.CdxV1RedeemTokenResponseResourcesOneOf{
				{
					CdxV1SharedTopic: &cdxv1.CdxV1SharedTopic{
						Kind:  "Topic",
						Topic: "topic-12345",
					},
				},
				{
					CdxV1SharedGroup: &cdxv1.CdxV1SharedGroup{
						Kind:        "Group",
						GroupPrefix: "stream-share.ss-12345",
					},
				},
			},
		}
		b, err := json.Marshal(&response)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for: "/cdx/v1/consumer-shared-resources"
func handleConsumerSharedResources(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list := &cdxv1.CdxV1ConsumerSharedResourceList{Data: []cdxv1.CdxV1ConsumerSharedResource{getTestConsumerSharedResource()}}
		b, err := json.Marshal(list)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for: "/cdx/v1/consumer-shared-resources/{id}:network"
func handlePrivateLinkNetworkConfig(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		network := cdxv1.CdxV1Network{
			DnsDomain:       stringToPtr("abc123.us-west-2.aws.stag.cpdev.cloud"),
			Zones:           &[]string{"usw2-az1", "usw2-az3", "usw2-az2"},
			ZonalSubdomains: &map[string]string{"usw2-az2": "usw2-az2.abc123.us-west-2.aws.stag.cpdev.cloud"},
			Cloud:           &cdxv1.CdxV1NetworkCloudOneOf{CdxV1AwsNetwork: getTestAWSNetwork()},
		}
		b, err := json.Marshal(&network)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for: "/cdx/v1/opt-in"
func handleOptInOptOut(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody cdxv1.CdxV1OptIn
		_ = json.Unmarshal(body, &reqBody)

		network := &cdxv1.CdxV1OptIn{StreamShareEnabled: reqBody.StreamShareEnabled}
		b, err := json.Marshal(&network)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}
