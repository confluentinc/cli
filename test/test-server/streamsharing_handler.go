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
		Id:               stringToPtr("ss-12345"),
		ProviderUserName: stringToPtr("provider"),
		Status:           stringToPtr("active"),
		SharedResource:   &cdxv1.ObjectReference{Id: "sr-12345"},
		InviteExpiresAt:  &expiresAt,
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
		ProviderUserName:         stringToPtr("provider"),
		Status:                   stringToPtr("active"),
		DeliveryMethod:           stringToPtr("email"),
		ServiceAccount:           &cdxv1.ObjectReference{Id: "sa-123456"},
		SharedResource:           &cdxv1.ObjectReference{Id: "sr-12345"},
		RedeemedAt:               &redeemedAt,
		InvitedAt:                &invitedAt,
		InviteExpiresAt:          &expiresAt,
	}
}

// Handler for: "/cdx/v1/provider-shares/{id}:resend"
func handleStreamSharingResendInvite(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for: "/cdx/v1/provider-shares"
func handleStreamSharingProviderShares(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
		response := cdxv1.CdxV1RedeemTokenResponse{
			Id:                stringPtr("ss-12345"),
			Apikey:            stringPtr("00000000000000000000"),
			Secret:            stringPtr("00000000000000000000"),
			KafkaBootstrapUrl: stringPtr("pkc-00000.us-east1.gcp.confluent.cloud:9092"),
			Resources: &[]cdxv1.CdxV1RedeemTokenResponseResourcesOneOf{
				{
					CdxV1SharedTopic: &cdxv1.CdxV1SharedTopic{
						Kind:  "Topic",
						Topic: "topic-12345",
					},
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
