package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
)

// Handler for: /cli/v1/feedbacks
func handleFeedbacks(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(cliv1.CliV1Feedback)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)
		if len(*req.Content) > 20 {
			w.WriteHeader(http.StatusForbidden)
			errorJson, err := json.Marshal(&struct {
				Errors []map[string]any `json:"errors"`
			}{
				Errors: []map[string]any{{"status": "403", "detail": "feedback exceeds the maximum length"}},
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(errorJson))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
