package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
)

var (
	invitedAt, _  = time.Parse(time.RFC3339, "2022-07-20T22:08:41+00:00")
	redeemedAt, _ = time.Parse(time.RFC3339, "2022-07-21T22:08:41+00:00")
	expiresAt, _  = time.Parse(time.RFC3339, "2022-07-22T22:08:41+00:00")

	consumerShares = []*cdxv1.CdxV1ConsumerShare{
		{
			Id:                       cdxv1.PtrString("ss-12345"),
			ProviderUserName:         cdxv1.PtrString("provider"),
			ProviderOrganizationName: cdxv1.PtrString("provider org"),
			Status:                   &cdxv1.CdxV1ConsumerShareStatus{Phase: "active"},
			InviteExpiresAt:          &expiresAt,
		},
		{
			Id:                       cdxv1.PtrString("ss-67890"),
			ProviderUserName:         cdxv1.PtrString("provider2"),
			ProviderOrganizationName: cdxv1.PtrString("provider org 2"),
			Status:                   &cdxv1.CdxV1ConsumerShareStatus{Phase: "active"},
			InviteExpiresAt:          &expiresAt,
		},
	}

	providerShares = []*cdxv1.CdxV1ProviderShare{
		{
			Id:                       cdxv1.PtrString("ss-12345"),
			ConsumerUserName:         cdxv1.PtrString("consumer"),
			ConsumerOrganizationName: cdxv1.PtrString("consumer org"),
			Status:                   &cdxv1.CdxV1ProviderShareStatus{Phase: "active"},
			DeliveryMethod:           cdxv1.PtrString("email"),
			RedeemedAt:               &redeemedAt,
			InvitedAt:                &invitedAt,
			InviteExpiresAt:          &expiresAt,
		},
		{
			Id:                       cdxv1.PtrString("ss-67890"),
			ConsumerUserName:         cdxv1.PtrString("consumer2"),
			ConsumerOrganizationName: cdxv1.PtrString("consumer org 2"),
			Status:                   &cdxv1.CdxV1ProviderShareStatus{Phase: "active"},
			DeliveryMethod:           cdxv1.PtrString("email"),
			RedeemedAt:               &redeemedAt,
			InvitedAt:                &invitedAt,
			InviteExpiresAt:          &expiresAt,
		},
	}
)

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
			err := json.NewEncoder(w).Encode(cdxv1.CdxV1ProviderShareList{Data: getV2List(providerShares)})
			require.NoError(t, err)
		case http.MethodPost:
			err := json.NewEncoder(w).Encode(providerShares[0])
			require.NoError(t, err)
		}
	}
}

// Handler for: "/cdx/v1/provider-shares/{id}"
func handleStreamSharingProviderShare(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		id = strings.TrimSuffix(id, ":resend")
		if i := getV2Index(providerShares, id); i != -1 {
			switch r.Method {
			case http.MethodGet:
				err := json.NewEncoder(w).Encode(providerShares[i])
				require.NoError(t, err)
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			}
		} else {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/cdx/v1/consumer-shares"
func handleStreamSharingConsumerShares(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(cdxv1.CdxV1ConsumerShareList{Data: getV2List(consumerShares)})
		require.NoError(t, err)
	}
}

// Handler for: "/cdx/v1/consumer-shares/{id}"
func handleStreamSharingConsumerShare(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if i := getV2Index(consumerShares, id); i != -1 {
			switch r.Method {
			case http.MethodGet:
				err := json.NewEncoder(w).Encode(consumerShares[i])
				require.NoError(t, err)
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			}
		} else {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/cdx/v1/shared-tokens:redeem"
func handleStreamSharingRedeemToken(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := cdxv1.CdxV1RedeemTokenResponse{
			Id:                   cdxv1.PtrString("ss-12345"),
			ApiKey:               cdxv1.PtrString("00000000000000000000"),
			Secret:               cdxv1.PtrString("00000000000000000000"),
			KafkaBootstrapUrl:    cdxv1.PtrString("pkc-00000.us-east1.gcp.confluent.cloud:9092"),
			SchemaRegistryUrl:    cdxv1.PtrString("https://psrc-xyz123.us-west-2.aws.cpdev.cloud"),
			SchemaRegistryApiKey: cdxv1.PtrString("00000000000000000000"),
			SchemaRegistrySecret: cdxv1.PtrString("00000000000000000000"),
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
		list := &cdxv1.CdxV1ConsumerSharedResourceList{Data: []cdxv1.CdxV1ConsumerSharedResource{{Id: cdxv1.PtrString("sr-12345")}}}
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
			DnsDomain:       cdxv1.PtrString("abc123.us-west-2.aws.stag.cpdev.cloud"),
			Zones:           &[]string{"usw2-az1", "usw2-az3", "usw2-az2"},
			ZonalSubdomains: &map[string]string{"usw2-az2": "usw2-az2.abc123.us-west-2.aws.stag.cpdev.cloud"},
			Cloud: &cdxv1.CdxV1NetworkCloudOneOf{
				CdxV1AwsNetwork: &cdxv1.CdxV1AwsNetwork{
					Kind:                       "AwsNetwork",
					PrivateLinkEndpointService: cdxv1.PtrString("com.amazonaws.vpce.us-west-2.vpce-svc-0000000000"),
				},
			},
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
