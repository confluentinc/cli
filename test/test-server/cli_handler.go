package testserver

import (
	"encoding/json"
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
		if len(req.GetContent()) > 20 {
			w.WriteHeader(http.StatusForbidden)
			err = writeErrorJson(w, "feedback exceeds the maximum length")
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
