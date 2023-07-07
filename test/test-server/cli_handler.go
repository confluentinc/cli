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
		feedback := cliv1.CliV1Feedback{Content: cliv1.PtrString("This CLI is great!")}
		b, err := json.Marshal(&feedback)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}
