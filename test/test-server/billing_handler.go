package testserver

import (
	"encoding/json"
	"fmt"
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
			var costList billingv1.BillingV1CostList

			// test over 10000 entries to ensure command does not stall terminal
			var data []billingv1.BillingV1Cost
			for i := 0; i < 10100; i++ {
				item := billingv1.BillingV1Cost{
					StartDate:   billingv1.PtrString("2023-01-01 00:00:00"),
					EndDate:     billingv1.PtrString("2023-01-02 00:00:00"),
					Granularity: billingv1.PtrString("DAILY"),
					LineType:    billingv1.PtrString("LINE_TYPE_1"),
					Product:     billingv1.PtrString("KAFKA"),
					Resource: &billingv1.BillingV1Resource{
						Id:          billingv1.PtrString(fmt.Sprintf("lkc-%d", i)),
						DisplayName: billingv1.PtrString("kafka_1"),
					},
					NetworkAccessType: billingv1.PtrString("INTERNET"),
					Price:             billingv1.PtrFloat64(0.123),
					Unit:              billingv1.PtrString("GB"),
					Amount:            billingv1.PtrFloat64(50.0),
					DiscountAmount:    billingv1.PtrFloat64(20.0),
					OriginalAmount:    billingv1.PtrFloat64(70.0),
				}

				data = append(data, item)
				env := billingv1.BillingV1Environment{Id: billingv1.PtrString("env-123")}
				item.Resource.SetEnvironment(env)
			}

			// test floats that could be 0
			nilFloatValues := billingv1.BillingV1Cost{
				StartDate:      billingv1.PtrString("2023-01-01 00:00:00"),
				EndDate:        billingv1.PtrString("2023-01-02 00:00:00"),
				Granularity:    billingv1.PtrString("DAILY"),
				LineType:       billingv1.PtrString("SUPPORT"),
				Product:        billingv1.PtrString("SUPPORT_PREMIER"),
				Price:          nil,
				Amount:         nil,
				DiscountAmount: nil,
				OriginalAmount: billingv1.PtrFloat64(70.0),
			}
			data = append(data, nilFloatValues)
			costList.Data = data
			err := json.NewEncoder(w).Encode(costList)
			require.NoError(t, err)
		}
	}
}
