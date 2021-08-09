package test_server

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/ccloud-sdk-go-v1"
)

// Handler for "/v2/metrics/cloud/query"
func (c *CloudRouter) HandleMetricsQuery(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		response := &ccloud.MetricsApiQueryReply{
			Result: []ccloud.ApiData{
				{
					Timestamp: time.Date(2019, 12, 19, 16, 1, 0, 0, time.UTC),
					Value:     1000,
					Labels:    map[string]interface{}{"metric.topic": "test-topic"},
				},
			},
		}
		w.Header().Add("Content-Type", "application/json")
		_, err := fmt.Fprint(w, response)
		require.NoError(t, err)
	}
}

func (c *CloudRouter) HandleAccessTokens(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_, err := fmt.Fprint(w, `{"token": "JWT_FROM_CCLOUD_API"}`)
		require.NoError(t, err)
	}
}
