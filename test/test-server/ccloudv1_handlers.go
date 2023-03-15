package testserver

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"
)

// Handler for "/api/organizations/"
func handlePriceTable(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prices := map[string]float64{
			strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleClusterType, exampleNetworkType}, ":"): examplePrice,
		}

		srPrices := map[string]float64{exampleSRPriceKey: exampleSchemaLimit}

		res := &ccloudv1.GetPriceTableReply{
			PriceTable: &ccloudv1.PriceTable{
				PriceTable: map[string]*ccloudv1.UnitPrices{
					exampleMetric:       {Unit: exampleUnit, Prices: prices},
					exampleSRPriceTable: {Unit: exampleSRPriceUnit, Prices: srPrices},
				},
			},
		}

		data, err := json.Marshal(res)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}

// Handler for: "/api/organizations/{id}/promo_code_claims"
func handlePromoCodeClaims(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			var tenDollars int64 = 10 * 10000

			// The time is set to noon so that all time zones display the same local time
			date := time.Date(2021, time.June, 16, 12, 0, 0, 0, time.UTC)
			expiration := &types.Timestamp{Seconds: date.Unix()}

			res := &ccloudv1.GetPromoCodeClaimsReply{
				Claims: []*ccloudv1.PromoCodeClaim{
					{
						Code:                 "PROMOCODE1",
						Amount:               tenDollars,
						Balance:              tenDollars,
						CreditExpirationDate: expiration,
					},
					{
						Code:                 "PROMOCODE2",
						Balance:              tenDollars,
						Amount:               tenDollars,
						CreditExpirationDate: expiration,
					},
				},
			}

			listReply, err := ccloudv1.MarshalJSONToBytes(res)
			require.NoError(t, err)
			_, err = w.Write(listReply)
			require.NoError(t, err)
		case http.MethodPost:
			res := &ccloudv1.ClaimPromoCodeReply{}

			err := json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/api/growth/v1/free-trial-info"
func handleFreeTrialInfo(t *testing.T) http.HandlerFunc {
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
