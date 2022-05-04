package testserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/ccloud-sdk-go-v1"
)

// Handler for "/v2/metrics/cloud/query"
func (c *CloudRouter) HandleMetricsQuery(t *testing.T) http.HandlerFunc {
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

func (c *CloudRouter) HandleJwtToken(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type CreateOAuthTokenReply struct {
			Token string `json:"token"`
		}
		err := json.NewEncoder(w).Encode(&CreateOAuthTokenReply{Token: "OAUTH_TOKEN"})
		require.NoError(t, err)
	}
}
