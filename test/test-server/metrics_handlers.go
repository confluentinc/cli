package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	"github.com/stretchr/testify/require"
)

func handleMetricsQuery(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("metrics being called..")
		w.Header().Set("Content-Type", "application/json")
		resp := &metricsv2.QueryResponse{
			FlatQueryResponse: &metricsv2.FlatQueryResponse{
				Data: []metricsv2.Point{
					{Value: 0.0, Timestamp: time.Date(2019, 12, 19, 16, 1, 0, 0, time.UTC)},
				},
			},
		}
		err := json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}
}
