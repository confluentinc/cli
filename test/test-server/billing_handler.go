package testserver

import (
	"encoding/json"
	billingv1 "github.com/confluentinc/ccloud-sdk-go-v2/billing/v1"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// Handler for: "/billing/v1/costs"
func handleBillingCosts(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			var costList = billingv1.BillingV1CostList{
				Data: []billingv1.BillingV1Cost{
					{
						StartDate:   billingv1.PtrString("2023-01-01 00:00:00"),
						EndDate:     billingv1.PtrString("2023-01-02 00:00:00"),
						Granularity: billingv1.PtrString("DAILY"),
						LineType:    billingv1.PtrString("LINE_TYPE_1"),
						Product:     billingv1.PtrString("KAFKA"),
						Resource: &billingv1.BillingV1Resource{
							Id:          billingv1.PtrString("lkc-123"),
							DisplayName: billingv1.PtrString("kafka_1"),
						},
						NetworkAccessType: billingv1.PtrString("INTERNET"),
						Price:             billingv1.PtrFloat64(0.123),
						Unit:              billingv1.PtrString("GB"),
						Amount:            billingv1.PtrFloat64(50.0),
						DiscountAmount:    billingv1.PtrFloat64(20.0),
						OriginalAmount:    billingv1.PtrFloat64(70.0),
					},
					{
						StartDate:   billingv1.PtrString("2023-01-01 00:00:00"),
						EndDate:     billingv1.PtrString("2023-02-03 00:00:00"),
						Granularity: billingv1.PtrString("DAILY"),
						LineType:    billingv1.PtrString("LINE_TYPE_2"),
						Product:     billingv1.PtrString("KAFKA"),
						Resource: &billingv1.BillingV1Resource{
							Id:          billingv1.PtrString("lkc-123"),
							DisplayName: billingv1.PtrString("kafka_4"),
						},
						NetworkAccessType: billingv1.PtrString("INTERNET"),
						Price:             billingv1.PtrFloat64(0.334),
						Unit:              billingv1.PtrString("GB"),
						Amount:            billingv1.PtrFloat64(50.0),
						DiscountAmount:    billingv1.PtrFloat64(20.0),
						OriginalAmount:    billingv1.PtrFloat64(70.0),
					},
				},
			}
			env := billingv1.BillingV1Environment{Id: billingv1.PtrString("env-123")}
			costList.Data[0].Resource.SetEnvironment(env)
			costList.Data[1].Resource.SetEnvironment(env)
			err := json.NewEncoder(w).Encode(costList)
			require.NoError(t, err)
		}
	}
}
