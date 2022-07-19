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

func handleStreamSharingListProviderShares(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		invitedAt, _ := time.Parse(time.RFC3339, "2022-07-19T15:04:05-07:00")
		redeemedAt, _ := time.Parse(time.RFC3339, "2022-07-20T15:04:05-07:00")
		expiresAt, _ := time.Parse(time.RFC3339, "2022-07-21T15:04:05-07:00")
		list := v1.CdxV1ProviderShareList{
			Data: []v1.CdxV1ProviderShare{
				{
					Id:                       stringToPtr("id"),
					ConsumerUserName:         stringToPtr("consumer"),
					ConsumerOrganizationName: stringToPtr("consumer org"),
					ProviderUserName:         stringToPtr("provider"),
					Status:                   stringToPtr("active"),
					DeliveryMethod:           stringToPtr("email"),
					ServiceAccount: &v1.ObjectReference{
						Id: "service account",
					},
					SharedResource: &v1.ObjectReference{
						Id: "shared resource",
					},
					RedeemedAt:      &redeemedAt,
					InvitedAt:       &invitedAt,
					InviteExpiresAt: &expiresAt,
				},
			},
		}
		b, err := json.Marshal(&list)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}
