package testserver

import (
	"encoding/json"
	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/cdx/v1"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

func getTestProviderShare() v1.CdxV1ProviderShare {
	invitedAt, _ := time.Parse(time.RFC3339, "2022-07-20T22:08:41+00:00")
	redeemedAt, _ := time.Parse(time.RFC3339, "2022-07-21T22:08:41+00:00")
	expiresAt, _ := time.Parse(time.RFC3339, "2022-07-22T22:08:41+00:00")
	return v1.CdxV1ProviderShare{
		Id:                       stringToPtr("id"),
		ConsumerUserName:         stringToPtr("consumer"),
		ConsumerOrganizationName: stringToPtr("consumer org"),
		ProviderUserName:         stringToPtr("provider"),
		Status:                   stringToPtr("active"),
		DeliveryMethod:           stringToPtr("email"),
		ServiceAccount: &v1.ObjectReference{
			Id: "sa-123456",
		},
		SharedResource: &v1.ObjectReference{
			Id: "shared resource",
		},
		RedeemedAt:      &redeemedAt,
		InvitedAt:       &invitedAt,
		InviteExpiresAt: &expiresAt,
	}
}

// Handler for: "/cdx/v1/provider-shares"
func handleStreamSharingProviderShares(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		list := v1.CdxV1ProviderShareList{
			Data: []v1.CdxV1ProviderShare{getTestProviderShare()},
		}
		b, err := json.Marshal(&list)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
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
