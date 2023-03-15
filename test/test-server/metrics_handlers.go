package testserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	"github.com/stretchr/testify/require"
)

var queryTime = time.Date(2019, 12, 19, 16, 1, 0, 0, time.UTC)

func handleMetricsQuery(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := &metricsv2.QueryResponse{
			FlatQueryResponse: &metricsv2.FlatQueryResponse{
				Data: []metricsv2.Point{
					{Value: 0.0, Timestamp: queryTime},
				},
			},
		}
		err := json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}
}
