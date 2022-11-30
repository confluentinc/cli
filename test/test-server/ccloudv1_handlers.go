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
func (c *CloudRouter) HandlePriceTable(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prices := map[string]float64{
			strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleClusterType, exampleNetworkType}, ":"): examplePrice,
		}

		res := &ccloudv1.GetPriceTableReply{
			PriceTable: &ccloudv1.PriceTable{
				PriceTable: map[string]*ccloudv1.UnitPrices{
					exampleMetric: {Unit: exampleUnit, Prices: prices},
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
func (c *CloudRouter) HandlePromoCodeClaims(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			var res *ccloudv1.GetPromoCodeClaimsReply

			var tenDollars int64 = 10 * 10000

			// The time is set to noon so that all time zones display the same local time
			date := time.Date(2021, time.June, 16, 12, 0, 0, 0, time.UTC)
			expiration := &types.Timestamp{Seconds: date.Unix()}

			freeTrialCode := &ccloudv1.GetPromoCodeClaimsReply{
				Claims: []*ccloudv1.PromoCodeClaim{
					{
						Code:                 PromoTestCode,
						Amount:               400 * 10000,
						Balance:              0,
						CreditExpirationDate: expiration,
					},
				},
			}

			regularCodes := &ccloudv1.GetPromoCodeClaimsReply{
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

			hasPromoCodeClaims := os.Getenv("HAS_PROMO_CODE_CLAIMS")
			switch hasPromoCodeClaims {
			case "false":
				res = &ccloudv1.GetPromoCodeClaimsReply{}
			case "onlyFreeTrialCode":
				res = freeTrialCode
			case "multiCodes":
				res = &ccloudv1.GetPromoCodeClaimsReply{}
				res.Claims = append(freeTrialCode.Claims, regularCodes.Claims...)
			default:
				res = regularCodes
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
