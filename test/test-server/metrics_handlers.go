package test_server

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"

	"github.com/confluentinc/ccloud-sdk-go-v1"
)

// Handler for "/v2/metrics/cloud/query"
func (c *CloudRouter) HandleMetricsQuery(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		response := &ccloud.MetricsApiQueryReply{
			Result: []ccloud.ApiData{
				{
					Timestamp: time.Date(2019, 12, 19, 16, 1, 0, 0, time.UTC),
					Value:     0,
					Labels:    map[string]interface{}{"metric.topic": "test-topic"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}
}
