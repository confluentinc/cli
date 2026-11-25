package httpclient

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserAgentTransport(t *testing.T) {
	t.Cleanup(func() { userAgent = "" })

	tests := []struct {
		name              string
		globalUserAgent   string
		existingUserAgent string
		want              []string
	}{
		{"sets when absent", "cli", "", []string{"cli"}},
		{"keeps an existing value, no duplicate", "cli", "sdk", []string{"sdk"}},
		{"passes through when unset", "", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetUserAgent(tt.globalUserAgent)

			req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
			require.NoError(t, err)
			if tt.existingUserAgent != "" {
				req.Header.Add("User-Agent", tt.existingUserAgent)
			}

			var actual []string
			next := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				actual = r.Header.Values("User-Agent")
				return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
			})
			_, err = (&userAgentTransport{next: next}).RoundTrip(req)
			require.NoError(t, err)

			require.Equal(t, tt.want, actual)
			// RoundTrip must not mutate the caller's request
			require.Equal(t, tt.existingUserAgent, req.Header.Get("User-Agent"))
		})
	}
}
