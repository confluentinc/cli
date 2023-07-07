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
		var req cliv1.CliV1Feedback
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		if len(*req.Content) > 20 {
			err := writeFeedbackExceedsMaxLengthError(w)
			require.NoError(t, err)
			return
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
