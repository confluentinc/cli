package testserver

import (
	"net/http"
	"os"
	"testing"
	"time"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"
)

// Handler for: "/api/growth/v1/free-trial-info"
func (c *CloudRouter) HandleFreeTrialInfo(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			date := time.Date(2021, time.June, 16, 12, 0, 0, 0, time.UTC)
			expiration := &types.Timestamp{Seconds: date.Unix()}

			var res *ccloudv1.GetFreeTrialInfoReply

			hasPromoCodeClaims := os.Getenv("IS_ON_FREE_TRIAL")
			switch hasPromoCodeClaims {
			case "true":
				res = &ccloudv1.GetFreeTrialInfoReply{
					PromoCodeClaims: []*ccloudv1.GrowthPromoCodeClaim{
						{
							Amount:               400 * 10000, // $400
							Balance:              20 * 10000,  // $20
							ClaimDate:            expiration,
							CreditExpirationDate: expiration,
							IsFreeTrialPromoCode: true,
						},
						{
							Amount:               20 * 10000, // $20
							Balance:              20 * 10000, // $20
							ClaimDate:            expiration,
							CreditExpirationDate: expiration,
							IsFreeTrialPromoCode: true,
						},
					},
				}
			default:
				res = &ccloudv1.GetFreeTrialInfoReply{}
			}

			reply, err := ccloudv1.MarshalJSONToBytes(res)
			require.NoError(t, err)
			_, err = w.Write(reply)
			require.NoError(t, err)
		}
	}
}
